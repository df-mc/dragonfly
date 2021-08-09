package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// BlockResult is the result of a ray trace collision with a block's model.
type BlockResult struct {
	bb   physics.AABB
	pos  mgl64.Vec3
	face cube.Face

	blockPos cube.Pos
}

// AABB returns the AABB that was collided within the block's model.
func (r BlockResult) AABB() physics.AABB {
	return r.bb
}

// Position ...
func (r BlockResult) Position() mgl64.Vec3 {
	return r.pos
}

// Face returns the hit block face.
func (r BlockResult) Face() cube.Face {
	return r.face
}

// BlockPosition ...
func (r BlockResult) BlockPosition() cube.Pos {
	return r.blockPos
}

// BlockIntercept performs a ray trace and calculates the point on the block model's edge nearest to the start position
// that the ray-trace collided with.
// BlockIntercept returns a BlockResult with the block collided with and with the colliding vector closest to the start position,
// if no colliding point was found, it returns nil.
func BlockIntercept(pos cube.Pos, w *world.World, b world.Block, pos1, pos2 mgl64.Vec3) Result {
	bbs := b.Model().AABB(pos, w)
	if len(bbs) == 0 {
		return nil
	}

	var (
		hit  Result
		dist = math.MaxFloat64
	)

	for _, bb := range bbs {
		next := Intercept(bb, pos1, pos2)
		if next == nil {
			continue
		}

		nextDist := next.Position().Sub(pos1).LenSqr()
		if nextDist < dist {
			hit = next
			dist = nextDist
		}
	}

	if hit == nil {
		return nil
	}

	return BlockResult{pos: hit.Position(), face: hit.Face(), blockPos: pos}
}

func (r BlockResult) __() {}
