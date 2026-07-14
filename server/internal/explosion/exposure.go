package explosion

import (
	"math"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Exposure returns the fraction of an entity's bounding box visible from an
// explosion origin.
func Exposure(tx *world.Tx, origin mgl64.Vec3, e world.Entity) float64 {
	box := e.H().Type().BBox(e).Translate(e.Position())
	boxMin, boxMax := box.Min(), box.Max()
	diff := boxMax.Sub(boxMin).Mul(2).Add(mgl64.Vec3{1, 1, 1})
	step := mgl64.Vec3{1 / diff[0], 1 / diff[1], 1 / diff[2]}
	if step[0] < 0 || step[1] < 0 || step[2] < 0 {
		return 0
	}

	xOffset := (1 - math.Floor(diff[0])/diff[0]) / 2
	zOffset := (1 - math.Floor(diff[2])/diff[2]) / 2
	var visible, total float64
	for x := 0.0; x <= 1; x += step[0] {
		for y := 0.0; y <= 1; y += step[1] {
			for z := 0.0; z <= 1; z += step[2] {
				point := mgl64.Vec3{
					lerp(boxMin[0], boxMax[0], x) + xOffset,
					lerp(boxMin[1], boxMax[1], y),
					lerp(boxMin[2], boxMax[2], z) + zOffset,
				}
				blocked := false
				trace.TraverseBlocks(origin, point, func(pos cube.Pos) bool {
					_, blocked = trace.BlockIntercept(pos, tx, tx.Block(pos), origin, point)
					return !blocked
				})
				if !blocked {
					visible++
				}
				total++
			}
		}
	}
	return visible / total
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}
