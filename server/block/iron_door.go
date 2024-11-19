package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// IronDoor is a variant of the door made of iron that can only be opened using redstone.
type IronDoor struct {
	transparent
	bass
	sourceWaterDisplacer

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
func (d IronDoor) Model() world.BlockModel {
	return model.Door{Facing: d.Facing, Open: d.Open, Right: d.Right}
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
	if _, ok := left.(IronDoor); ok {
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
	place(w, pos.Side(cube.FaceUp), IronDoor{Facing: d.Facing, Top: true, Right: d.Right}, user, ctx)
	ctx.SubtractFromCount(1)
	return placed(ctx)
}

// BreakInfo ...
func (d IronDoor) BreakInfo() BreakInfo {
	return newBreakInfo(3, alwaysHarvestable, axeEffective, oneOf(d))
}

// SideClosed ...
func (d IronDoor) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// RedstoneUpdate ...
func (d IronDoor) RedstoneUpdate(pos cube.Pos, w *world.World) {
	if d.Open == receivedRedstonePower(pos, w) {
		return
	}
	if !d.Open {
		d.Open = true
		w.PlaySound(pos.Vec3Centre(), sound.DoorOpen{Block: d})
		w.SetBlock(pos, d, &world.SetOpts{DisableBlockUpdates: true})
	} else {
		w.ScheduleBlockUpdate(pos, time.Millisecond*200)
	}
}

// ScheduledTick ...
func (d IronDoor) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if receivedRedstonePower(pos, w) {
		return
	}
	d.Open = false
	w.PlaySound(pos.Vec3Centre(), sound.DoorClose{Block: d})
	w.SetBlock(pos, d, &world.SetOpts{DisableBlockUpdates: true})
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

// allIronDoors returns a list of all door types
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
