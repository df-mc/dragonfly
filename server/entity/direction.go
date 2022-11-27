package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Eyed represents an entity that has eyes.
type Eyed interface {
	// EyeHeight returns the offset from their base position that the eyes of an entity are found at.
	EyeHeight() float64
}

// EyePosition returns the position of the eyes of the entity if the entity implements entity.Eyed, or the
// actual position if it doesn't.
func EyePosition(e world.Entity) mgl64.Vec3 {
	pos := e.Position()
	if eyed, ok := e.(Eyed); ok {
		pos[1] += eyed.EyeHeight()
	}
	return pos
}
