package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// doorPlacementPosition keeps a replaceable clicked cell or uses a solid block's top face.
func doorPlacementPosition(pos cube.Pos, face cube.Face, tx *world.Tx, door world.Block) (cube.Pos, bool) {
	var used bool
	pos, face, used = firstReplaceable(tx, pos, face, door)
	if !used || face != cube.FaceUp || !replaceableWith(tx, pos.Side(cube.FaceUp), door) {
		return pos, false
	}

	below := pos.Side(cube.FaceDown)
	return pos, tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx)
}

// doorHingeRight chooses a hinge from doors and solid blocks beside both halves.
func doorHingeRight(pos cube.Pos, facing cube.Direction, tx *world.Tx) bool {
	left := pos.Side(facing.RotateLeft().Face())
	right := pos.Side(facing.RotateRight().Face())
	leftCount, leftDoor := doorHingeSide(left, tx)
	rightCount, rightDoor := doorHingeSide(right, tx)
	return leftDoor && !rightDoor || rightCount > leftCount
}

// doorHingeSide counts opaque full blocks and detects doors beside either half.
func doorHingeSide(pos cube.Pos, tx *world.Tx) (solid int, door bool) {
	for _, p := range []cube.Pos{pos, pos.Side(cube.FaceUp)} {
		b := tx.Block(p)
		if _, ok := b.Model().(model.Door); ok {
			door = true
		}
		if diffuser, ok := b.(LightDiffuser); ok && diffuser.LightDiffusionLevel() != 15 {
			continue
		}
		full := true
		for _, face := range cube.Faces() {
			if !b.Model().FaceSolid(p, face, tx) {
				full = false
				break
			}
		}
		if full {
			solid++
		}
	}
	return
}
