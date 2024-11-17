package block

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
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
	// Facing is the direction the door is facing.
	Facing cube.Direction
	// Open is whether the door is open.
	Open bool
	// Top is whether the block is the top or bottom half of a door
	Top bool
	// Right is whether the door hinge is on the right side
	Right bool
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

func (d CopperDoor) WithOxidationLevel(o OxidationType) Oxidizable {
	d.Oxidation = o
	return d
}

// NeighbourUpdateTick ...
func (d CopperDoor) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, w *world.World) {
	if pos == changedNeighbour {
		return
	}
	if d.Top {
		if b, ok := w.Block(pos.Side(cube.FaceDown)).(CopperDoor); !ok {
			w.SetBlock(pos, nil, nil)
			w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
		} else if d.Oxidation != b.Oxidation || d.Waxed != b.Waxed {
			d.Oxidation = b.Oxidation
			d.Waxed = b.Waxed
			fmt.Println("NeighbourUpdateTick 1", d, b)
			w.SetBlock(pos, d, nil)
		}
		return
	}
	if solid := w.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, w); !solid {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
	} else if b, ok := w.Block(pos.Side(cube.FaceUp)).(CopperDoor); !ok {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
	} else if d.Oxidation != b.Oxidation || d.Waxed != b.Waxed {
		d.Oxidation = b.Oxidation
		d.Waxed = b.Waxed
		fmt.Println("NeighbourUpdateTick 2", d, b)
		w.SetBlock(pos, d, nil)
	}
}

// UseOnBlock handles the directional placing of doors
func (d CopperDoor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	if face != cube.FaceUp {
		// Doors can only be placed when clicking the top face.
		return false
	}
	below := pos
	pos = pos.Side(cube.FaceUp)
	if !replaceableWith(w, pos, d) || !replaceableWith(w, pos.Side(cube.FaceUp), d) {
		return false
	}
	if !w.Block(below).Model().FaceSolid(below, cube.FaceUp, w) {
		return false
	}
	d.Facing = user.Rotation().Direction()
	left := w.Block(pos.Side(d.Facing.RotateLeft().Face()))
	right := w.Block(pos.Side(d.Facing.RotateRight().Face()))
	if _, ok := left.(CopperDoor); ok {
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
	place(w, pos, d, user, ctx)
	place(w, pos.Side(cube.FaceUp), CopperDoor{Oxidation: d.Oxidation, Waxed: d.Waxed, Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	ctx.SubtractFromCount(1)
	return placed(ctx)
}

func (d CopperDoor) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User, _ *item.UseContext) bool {
	d.Open = !d.Open
	w.SetBlock(pos, d, nil)

	otherPos := pos.Side(cube.Face(boolByte(!d.Top)))
	other := w.Block(otherPos)
	if door, ok := other.(CopperDoor); ok {
		door.Open = d.Open
		w.SetBlock(otherPos, door, nil)
	}
	if d.Open {
		w.PlaySound(pos.Vec3Centre(), sound.DoorOpen{Block: d})
		return true
	}
	w.PlaySound(pos.Vec3Centre(), sound.DoorClose{Block: d})
	return true
}

func (d CopperDoor) SneakingActivate(pos cube.Pos, _ cube.Face, w *world.World, user item.User, _ *item.UseContext) bool {
	var ok bool
	d.Oxidation, d.Waxed, ok = activateOxidizable(pos, w, user, d.Oxidation, d.Waxed)
	if ok {
		fmt.Println("SneakingActivate", d)
		w.SetBlock(pos, d, nil)
		return true
	}
	return false
}

func (d CopperDoor) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	attemptOxidation(pos, w, r, d)
}

// BreakInfo ...
func (d CopperDoor) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, axeEffective, oneOf(d))
}

// SideClosed ...
func (d CopperDoor) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (d CopperDoor) EncodeItem() (name string, meta int16) {
	name = "copper_door"
	if d.Oxidation != NormalOxidation() {
		name = d.Oxidation.String() + "_" + name
	}
	if d.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, 0
}

// EncodeBlock ...
func (d CopperDoor) EncodeBlock() (name string, properties map[string]any) {
	direction := 3
	switch d.Facing {
	case cube.South:
		direction = 1
	case cube.West:
		direction = 2
	case cube.East:
		direction = 0
	}

	name = "copper_door"
	if d.Oxidation != NormalOxidation() {
		name = d.Oxidation.String() + "_" + name
	}
	if d.Waxed {
		name = "waxed_" + name
	}
	return "minecraft:" + name, map[string]any{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
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
