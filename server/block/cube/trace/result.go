package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// Result represents the result of a ray trace collision with a bounding box.
type Result interface {
	// BBox returns the bounding box collided with.
	BBox() cube.BBox
	// Position returns where the ray first collided with the bounding box.
	Position() mgl64.Vec3
	// Face returns the face of the bounding box that was collided on.
	Face() cube.Face
}

// Perform performs a ray trace between start and end, checking if any blocks or entities collided with the
// ray. The physics.BBox that's passed is used for checking if any entity within the bounding box collided
// with the ray.
func Perform(start, end mgl64.Vec3, w *world.World, box cube.BBox, ignored func(world.Entity) bool) (hit Result, ok bool) {
	// Check if there's any blocks that we may collide with.
	TraverseBlocks(start, end, func(pos cube.Pos) (cont bool) {
		b := w.Block(pos)

		// Check if we collide with the block's model.
		if result, ok := BlockIntercept(pos, w, b, start, end); ok {
			hit = result
			end = hit.Position()
			return false
		}
		return true
	})

	// Now check for any entities that we may collide with.
	dist := math.MaxFloat64
	bb := box.Translate(start).Extend(end.Sub(start))
	for _, entity := range w.EntitiesWithin(bb.Grow(8.0), ignored) {
		if ignored != nil && ignored(entity) || !entity.BBox().Translate(entity.Position()).IntersectsWith(bb) {
			continue
		}
		// Check if we collide with the entities bounding box.
		result, ok := EntityIntercept(entity, start, end)
		if !ok {
			continue
		}

		if distance := start.Sub(result.Position()).LenSqr(); distance < dist {
			dist = distance
			hit = result
		}
	}

	return hit, hit != nil
}
