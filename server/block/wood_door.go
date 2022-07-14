package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodDoor is a block that can be used as an openable 1x2 barrier.
type WoodDoor struct {
	transparent
	bass

	// Wood is the type of wood of the door. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Facing is the direction the door is facing.
	Facing cube.Direction
	// Open is whether the door is open.
	Open bool
	// Top is whether the block is the top or bottom half of a door
	Top bool
	// Right is whether the door hinge is on the right side
	Right bool
}

// FlammabilityInfo ...
func (d WoodDoor) FlammabilityInfo() FlammabilityInfo {
	if !d.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(0, 0, true)
}

// Model ...
func (d WoodDoor) Model() world.BlockModel {
	return model.Door{Facing: d.Facing, Open: d.Open, Right: d.Right}
}

// NeighbourUpdateTick ...
func (d WoodDoor) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d.Top {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(WoodDoor); !ok {
			w.SetBlock(pos, nil, nil)
			w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
		}
		return
	}
	if solid := w.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, w); !solid {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
	} else if _, ok := w.Block(pos.Side(cube.FaceUp)).(WoodDoor); !ok {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
	}
}

// UseOnBlock handles the directional placing of doors
func (d WoodDoor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	if face != cube.FaceUp {
		// Doors can only be placed when clicking the top face.
		return false
	}
	below := pos
	pos = pos.Add(cube.Pos{0, 1})
	if !replaceableWith(w, pos, d) || !replaceableWith(w, pos.Side(cube.FaceUp), d) {
		return false
	}
	if !w.Block(below).Model().FaceSolid(below, cube.FaceUp, w) {
		return false
	}
	d.Facing = user.Facing()
	left := w.Block(pos.Side(d.Facing.RotateLeft().Face()))
	right := w.Block(pos.Side(d.Facing.RotateRight().Face()))
	if door, ok := left.(WoodDoor); ok {
		if door.Wood == d.Wood {
			d.Right = true
		}
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
	place(w, pos.Side(cube.FaceUp), WoodDoor{Wood: d.Wood, Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	ctx.CountSub = 1
	return placed(ctx)
}

// Activate ...
func (d WoodDoor) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User) bool {
	d.Open = !d.Open
	w.SetBlock(pos, d, nil)

	otherPos := pos.Side(cube.Face(boolByte(!d.Top)))
	other := w.Block(otherPos)
	if door, ok := other.(WoodDoor); ok {
		door.Open = d.Open
		w.SetBlock(otherPos, door, nil)
	}

	w.PlaySound(pos.Vec3Centre(), sound.Door{})
	return true
}

// BreakInfo ...
func (d WoodDoor) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, axeEffective, oneOf(d))
}

// CanDisplace ...
func (d WoodDoor) CanDisplace(l world.Liquid) bool {
	_, water := l.(Water)
	return water
}

// SideClosed ...
func (d WoodDoor) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (d WoodDoor) EncodeItem() (name string, meta int16) {
	if d.Wood == OakWood() {
		return "minecraft:wooden_door", 0
	}
	return "minecraft:" + d.Wood.String() + "_door", 0
}

// EncodeBlock ...
func (d WoodDoor) EncodeBlock() (name string, properties map[string]any) {
	direction := 3
	switch d.Facing {
	case cube.South:
		direction = 1
	case cube.West:
		direction = 2
	case cube.East:
		direction = 0
	}

	if d.Wood == OakWood() {
		return "minecraft:wooden_door", map[string]any{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
	}
	return "minecraft:" + d.Wood.String() + "_door", map[string]any{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
}

// allDoors returns a list of all door types
func allDoors() (doors []world.Block) {
	for _, w := range WoodTypes() {
		for i := cube.Direction(0); i <= 3; i++ {
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: false, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: true, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: true, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: false, Right: false})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: false, Right: true})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: false, Top: true, Right: true})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: true, Right: true})
			doors = append(doors, WoodDoor{Wood: w, Facing: i, Open: true, Top: false, Right: true})
		}
	}
	return
}
