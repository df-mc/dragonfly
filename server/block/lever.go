package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Lever is a non-solid block that can provide switchable redstone power.
type Lever struct {
	empty
	transparent
	flowingWaterDisplacer

	// Powered is if the lever is switched on.
	Powered bool
	// Facing is the face of the block that the lever is attached to.
	Facing cube.Face
	// Direction is the direction the lever is pointing. This is only used for levers that are attached on up or down
	// faces. Currently, only North and West directions are supported due to Bedrock Edition limitations.
	Direction cube.Direction
}

// RedstonePower returns maximum power while the lever is active.
func (l Lever) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if l.Powered {
		return 15
	}
	return 0
}

// RedstoneStrongPower strongly powers the block the lever is attached to.
func (l Lever) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if l.Powered && l.Facing.Opposite() == face {
		return 15
	}
	return 0
}

// SideClosed ...
func (l Lever) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// NeighbourUpdateTick ...
func (l Lever) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportPos := pos.Side(l.Facing.Opposite())
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, l.Facing, tx) {
		breakBlock(l, pos, tx)
	}
}

// UseOnBlock ...
func (l Lever) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, l)
	if !used {
		return false
	}
	supportPos := pos.Side(face.Opposite())
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, face, tx) {
		return false
	}

	l.Powered = false
	l.Facing = face
	l.Direction = cube.North
	if face.Axis() == cube.Y && user.Rotation().Direction().Face().Axis() == cube.X {
		l.Direction = cube.West
	}
	place(tx, pos, l, user, ctx)
	return placed(ctx)
}

// Activate ...
func (l Lever) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	l.Powered = !l.Powered
	tx.SetBlock(pos, l, nil)
	if l.Powered {
		tx.PlaySound(pos.Vec3Centre(), sound.PowerOn{})
	} else {
		tx.PlaySound(pos.Vec3Centre(), sound.PowerOff{})
	}
	return true
}

// BreakInfo ...
func (l Lever) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, nothingEffective, oneOf(Lever{})).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		tx.ScheduleRedstoneUpdate(pos)
	})
}

// EncodeItem ...
func (l Lever) EncodeItem() (name string, meta int16) {
	return "minecraft:lever", 0
}

// EncodeBlock ...
func (l Lever) EncodeBlock() (string, map[string]any) {
	direction := l.Facing.String()
	if l.Facing == cube.FaceDown || l.Facing == cube.FaceUp {
		axis := "east_west"
		if l.Direction == cube.North {
			axis = "north_south"
		}
		direction += "_" + axis
	}
	return "minecraft:lever", map[string]any{"open_bit": l.Powered, "lever_direction": direction}
}

// allLevers ...
func allLevers() (all []world.Block) {
	f := func(facing cube.Face, direction cube.Direction) {
		all = append(all, Lever{Facing: facing, Direction: direction})
		all = append(all, Lever{Facing: facing, Direction: direction, Powered: true})
	}
	for _, facing := range cube.Faces() {
		f(facing, cube.North)
		if facing == cube.FaceDown || facing == cube.FaceUp {
			f(facing, cube.West)
		}
	}
	return
}
