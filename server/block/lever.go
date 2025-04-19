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
	// faces.
	// TODO: Better handle lever direction on up or down faces—using a `cube.Axis` results in a default `Lever` with an
	//  axis `Y` and a face `Down` which does not map to an existing block state.
	Direction cube.Direction
}

// RedstoneSource ...
func (l Lever) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (l Lever) WeakPower(cube.Pos, cube.Face, *world.Tx, bool) int {
	if l.Powered {
		return 15
	}
	return 0
}

// StrongPower ...
func (l Lever) StrongPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if l.Powered && l.Facing == face {
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
	if !tx.Block(pos.Side(l.Facing.Opposite())).Model().FaceSolid(pos.Side(l.Facing.Opposite()), l.Facing, tx) {
		breakBlock(l, pos, tx)
		updateDirectionalRedstone(pos, tx, l.Facing.Opposite())
	}
}

// UseOnBlock ...
func (l Lever) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, l)
	if !used {
		return false
	}
	if !tx.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, tx) {
		return false
	}

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
	updateDirectionalRedstone(pos, tx, l.Facing.Opposite())
	return true
}

// BreakInfo ...
func (l Lever) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, nothingEffective, oneOf(l)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateDirectionalRedstone(pos, tx, l.Facing.Opposite())
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
