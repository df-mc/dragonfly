package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/go-gl/mathgl/mgl64"
)

// Result represents the result of a ray-trace collision with a bounding box.
type Result interface {
	// AABB returns the bounding box collided with.
	AABB() physics.AABB
	// Position ...
	Position() mgl64.Vec3
	// Face ...
	Face() cube.Face

	__()
}
