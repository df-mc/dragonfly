package entity

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// pushSpeed is how fast an entity inside a block is pushed out of it, in
// blocks per tick. Measured against BDS.
const pushSpeed = 0.21

// pushFaces are the faces an entity may be pushed out through.
var pushFaces = [...]cube.Face{cube.FaceNorth, cube.FaceSouth, cube.FaceWest, cube.FaceEast, cube.FaceUp}

var fullBox = cube.Box(0, 0, 0, 1, 1, 1)

// pushOutOfBlock moves an entity that ended up inside a block out of it
// through the nearest free side and returns the movement doing so, or nil if
// it is not in one. It replaces the entity's physics for the tick, as
// collision alone would hold the entity at the first block it is not yet
// inside.
func pushOutOfBlock(e *Ent, tx *world.Tx) *Movement {
	pos := e.Position()
	box := e.H().Type().BBox(e).Translate(pos)
	if !obstructed(tx, box) {
		return nil
	}
	centre := mgl64.Vec3{pos[0], (box.Min()[1] + box.Max()[1]) / 2, pos[2]}
	push, prevVel := pushDirection(tx, centre).Mul(pushSpeed), e.data.Vel
	return &Movement{v: tx.Viewers(pos), e: e,
		pos: pos.Add(push), vel: push, dpos: push, dvel: push.Sub(prevVel),
		rot: e.data.Rot, onGround: false,
	}
}

// pushDirection returns a unit vector towards the nearest face of the block
// at centre that is not blocked by a full block, or up if all of them are.
func pushDirection(tx *world.Tx, centre mgl64.Vec3) mgl64.Vec3 {
	blockPos := cube.PosFromVec3(centre)
	offset := centre.Sub(blockPos.Vec3())

	dir, nearest := cube.FaceUp.Axis().Vec3(), math.MaxFloat64
	for _, face := range pushFaces {
		if fullBlock(tx, blockPos.Side(face)) {
			continue
		}
		v := blockPos.Side(face).Sub(blockPos).Vec3()
		dist := v.Dot(offset)
		if v.X()+v.Y()+v.Z() > 0 {
			dist = 1 - dist
		} else {
			dist = -dist
		}
		if dist < nearest {
			dir, nearest = v, dist
		}
	}
	return dir
}

// fullBlock reports whether the block at pos has a full collision box.
func fullBlock(tx *world.Tx, pos cube.Pos) bool {
	boxes := tx.Block(pos).Model().BBox(pos, tx)
	return len(boxes) == 1 && boxes[0] == fullBox
}

// obstructed reports whether any block intersects the bounding box passed.
func obstructed(tx *world.Tx, box cube.BBox) bool {
	for pos := range cube.Range3D(cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())) {
		for _, blockBox := range tx.Block(pos).Model().BBox(pos, tx) {
			if blockBox.Translate(pos.Vec3()).IntersectsWith(box) {
				return true
			}
		}
	}
	return false
}
