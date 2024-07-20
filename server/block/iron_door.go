package block

import (
	"math/rand"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// IronDoor is a block that can be used as an openable 1x2 barrier.
type IronDoor struct {
	transparent
	bass
	sourceWaterDisplacer

	// Powered is whether the dropper is powered or not.
	Powered bool
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
func (d IronDoor) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// Model ...
func (d IronDoor) Model() world.BlockModel {
	return model.Door{Facing: d.Facing, Open: d.Open, Right: d.Right}
}

// PistonBreakable ...
func (IronDoor) PistonBreakable() bool {
	return true
}

// NeighbourUpdateTick ...
func (d IronDoor) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d.Top {
		if _, ok := w.Block(pos.Side(cube.FaceDown)).(IronDoor); !ok {
			w.SetBlock(pos, nil, nil)
			w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
		}
		return
	}
	if solid := w.Block(pos.Side(cube.FaceDown)).Model().FaceSolid(pos.Side(cube.FaceDown), cube.FaceUp, w); !solid {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
	} else if _, ok := w.Block(pos.Side(cube.FaceUp)).(IronDoor); !ok {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: d})
	}
}

// UseOnBlock handles the directional placing of doors
func (d IronDoor) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
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
	place(w, pos.Side(cube.FaceUp), IronDoor{Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	ctx.CountSub = 1
	return placed(ctx)
}

// Activate ...
func (d IronDoor) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User, _ *item.UseContext) bool {
	return true
}

// RedstoneUpdate ...
func (d IronDoor) RedstoneUpdate(pos cube.Pos, w *world.World) {
	powered := receivedRedstonePower(pos, w)
	if powered {
		d.Powered = true
		d.Open = true
	} else {
		d.Powered = false
		d.Open = false
	}

	w.SetBlock(pos, d, nil)
	w.ScheduleBlockUpdate(pos, time.Millisecond*200)

}

// ScheduledTick ...
func (d IronDoor) ScheduledTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if d.Powered {
		d.Open = true
		w.SetBlock(pos, d, nil)
	}
}

// BreakInfo ...
func (d IronDoor) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, axeEffective, oneOf(d))
}

// SideClosed ...
func (d IronDoor) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (d IronDoor) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_door", 0
}

// EncodeBlock ...
func (d IronDoor) EncodeBlock() (name string, properties map[string]any) {
	direction := 3
	switch d.Facing {
	case cube.South:
		direction = 1
	case cube.West:
		direction = 2
	case cube.East:
		direction = 0
	}

	return "minecraft:iron_door", map[string]any{"direction": int32(direction), "door_hinge_bit": d.Right, "open_bit": d.Open, "upper_block_bit": d.Top}
}

// allDoors returns a list of all door types
func allIronDoors() (doors []world.Block) {
	for i := cube.Direction(0); i <= 3; i++ {
		doors = append(doors, IronDoor{Facing: i, Open: false, Top: false, Right: false})
		doors = append(doors, IronDoor{Facing: i, Open: false, Top: true, Right: false})
		doors = append(doors, IronDoor{Facing: i, Open: true, Top: true, Right: false})
		doors = append(doors, IronDoor{Facing: i, Open: true, Top: false, Right: false})
		doors = append(doors, IronDoor{Facing: i, Open: false, Top: false, Right: true})
		doors = append(doors, IronDoor{Facing: i, Open: false, Top: true, Right: true})
		doors = append(doors, IronDoor{Facing: i, Open: true, Top: true, Right: true})
		doors = append(doors, IronDoor{Facing: i, Open: true, Top: false, Right: true})
	}
	return
}