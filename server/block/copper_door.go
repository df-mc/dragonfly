package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// CopperDoor is a block that can be used as an openable 1x2 barrier.
type CopperDoor struct {
	transparent
	bass
	sourceWaterDisplacer

	// Oxidation is the level of oxidation of the copper door.
	Oxidation OxidationType
	// Waxed bool is whether the copper door has been waxed with honeycomb.
	Waxed bool
	// Facing is the direction that the door opens towards. When closed, the door sits on the side of its
	// block on the opposite direction.
	Facing cube.Direction
	// Open is whether the door is open.
	Open bool
	// Top is whether the block is the top or bottom half of a door.
	Top bool
	// Right is whether the door hinge is on the right side.
	Right bool
}

func (d CopperDoor) Strip() (world.Block, world.Sound, bool) {
	if d.Waxed {
		d.Waxed = false
		return d, sound.WaxRemoved{}, true
	} else if ot, ok := d.Oxidation.Decrease(); ok {
		d.Oxidation = ot
		return d, sound.CopperScraped{}, true
	}
	return d, nil, false
}

// Model ...
func (d CopperDoor) Model() world.BlockModel {
	return model.Door{Facing: d.Facing, Open: d.Open, Right: d.Right}
}

// Wax waxes the copper door to stop it from oxidising further.
func (d CopperDoor) Wax(cube.Pos, mgl64.Vec3) (world.Block, bool) {
	if d.Waxed {
		return d, false
	}
	d.Waxed = true
	return d, true
}

func (d CopperDoor) CanOxidate() bool {
	return !d.Waxed
}

func (d CopperDoor) OxidationLevel() OxidationType {
	return d.Oxidation
}

func (d CopperDoor) WithOxidationLevel(o OxidationType) Oxidisable {
	d.Oxidation = o
	return d
}

// NeighbourUpdateTick ...
func (d CopperDoor) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *world.Tx) {
	if pos == changedNeighbour {
		return
	}
	if d.Top {
		if b, ok := tx.Block(pos.Side(cube.FaceDown)).(CopperDoor); !ok {
			breakBlockNoDrops(d, pos, tx)
		} else if d.Oxidation != b.Oxidation || d.Waxed != b.Waxed {
			d.Oxidation = b.Oxidation
			d.Waxed = b.Waxed
			tx.SetBlock(pos, d, nil)
		}
	} else if solid := tx.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, tx); !solid {
		// CopperDoor is pickaxeHarvestable, so don't use breakBlock() here.
		breakBlockNoDrops(d, pos, tx)
		dropItem(tx, item.NewStack(d, 1), pos.Vec3Centre())
	} else if b, ok := tx.Block(pos.Side(cube.FaceUp)).(CopperDoor); !ok {
		breakBlockNoDrops(d, pos, tx)
	} else if d.Oxidation != b.Oxidation || d.Waxed != b.Waxed {
		d.Oxidation = b.Oxidation
		d.Waxed = b.Waxed
		tx.SetBlock(pos, d, nil)
	}
}

// UseOnBlock handles the directional placing of doors
func (d CopperDoor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	if face != cube.FaceUp {
		// Doors can only be placed when clicking the top face.
		return false
	}
	below := pos
	pos = pos.Side(cube.FaceUp)
	if !replaceableWith(tx, pos, d) || !replaceableWith(tx, pos.Side(cube.FaceUp), d) {
		return false
	}
	if !tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx) {
		return false
	}
	d.Facing = user.Rotation().Direction()
	left := tx.Block(pos.Side(d.Facing.RotateLeft().Face()))
	right := tx.Block(pos.Side(d.Facing.RotateRight().Face()))
	if _, ok := left.Model().(model.Door); ok {
		d.Right = true
	}
	// The side the door hinge is on can be affected by the blocks to the left and right of the door. In particular,
	// opaque blocks on the right side of the door with transparent blocks on the left side result in a right sided
	// door hinge.
	if diffuser, ok := right.(LightDiffuser); !ok || diffuser.LightDiffusionLevel() != 0 {
		if diffuser, ok := left.(LightDiffuser); ok && diffuser.LightDiffusionLevel() == 0 {
			d.Right = true
		}
	}

	ctx.IgnoreBBox = true
	place(tx, pos, d, user, ctx)
	place(tx, pos.Side(cube.FaceUp), CopperDoor{Oxidation: d.Oxidation, Waxed: d.Waxed, Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	ctx.CountSub = 1
	return placed(ctx)
}

func (d CopperDoor) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	d.Open = !d.Open
	tx.SetBlock(pos, d, nil)

	otherPos := pos.Side(cube.Face(boolByte(!d.Top)))
	other := tx.Block(otherPos)
	if door, ok := other.(CopperDoor); ok {
		door.Open = d.Open
		tx.SetBlock(otherPos, door, nil)
	}
	if d.Open {
		tx.PlaySound(pos.Vec3Centre(), sound.DoorOpen{Block: d})
		return true
	}
	tx.PlaySound(pos.Vec3Centre(), sound.DoorClose{Block: d})
	return true
}

func (d CopperDoor) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	attemptOxidation(pos, tx, r, d)
}

// BreakInfo ...
func (d CopperDoor) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, oneOf(d))
}

// SideClosed ...
func (d CopperDoor) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeItem ...
func (d CopperDoor) EncodeItem() (name string, meta int16) {
	return copperBlockName("copper_door", d.Oxidation, d.Waxed), 0
}

// EncodeBlock ...
func (d CopperDoor) EncodeBlock() (name string, properties map[string]any) {
	return copperBlockName("copper_door", d.Oxidation, d.Waxed), map[string]any{"minecraft:cardinal_direction": d.Facing.RotateRight().String(), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
}

// allCopperDoors returns a list of all copper door types
func allCopperDoors() (doors []world.Block) {
	f := func(waxed bool) {
		for _, o := range OxidationTypes() {
			for i := cube.Direction(0); i <= 3; i++ {
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: false, Top: false, Right: false})
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: false, Top: true, Right: false})
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: true, Top: true, Right: false})
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: true, Top: false, Right: false})
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: false, Top: false, Right: true})
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: false, Top: true, Right: true})
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: true, Top: true, Right: true})
				doors = append(doors, CopperDoor{Oxidation: o, Waxed: waxed, Facing: i, Open: true, Top: false, Right: true})
			}
		}
	}
	f(false)
	f(true)
	return
}
