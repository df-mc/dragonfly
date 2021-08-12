package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/go-gl/mathgl/mgl64"
)

// Result represents the result of a ray trace collision with a bounding box.
type Result interface {
	// AABB returns the bounding box collided with.
	AABB() physics.AABB
	// Position returns where the ray first collided with the bounding box.
	Position() mgl64.Vec3
	// Face returns the face of the bounding box that was collided on.
	Face() cube.Face
}
