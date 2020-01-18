package particle

// Particle represents a particle that may be added to the world. These particles are then rendered client-
// side, with the server having no control over it after sending.
type Particle interface {
	__()
}
