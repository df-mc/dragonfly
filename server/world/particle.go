package world

import (
	"github.com/go-gl/mathgl/mgl64"
)

// Particle represents a particle that may be added to the world. These particles are then rendered client-
// side, with the server having no control over it after sending.
type Particle interface {
	// Spawn spawns the particle at the position passed. Particles may execute any additional actions here,
	// such as spawning different particles.
	Spawn(w *World, pos mgl64.Vec3)
}
