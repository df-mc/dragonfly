package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// BlockResult ...
type BlockResult struct {
	pos  mgl64.Vec3
	face cube.Face

	blockPos cube.Pos
}

// Position ...
func (r BlockResult) Position() mgl64.Vec3 {
	return r.pos
}

// Face ...
func (r BlockResult) Face() cube.Face {
	return r.face
}

// BlockPosition ...
func (r BlockResult) BlockPosition() cube.Pos {
	return r.blockPos
}

// BlockIntercept ...
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
