package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

// Result ...
type Result interface {
	// Position ...
	Position() mgl64.Vec3
	// Face ...
	Face() cube.Face

	__()
}
