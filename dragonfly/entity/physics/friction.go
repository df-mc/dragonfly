package physics

// DefaultFriction is the default friction that blocks have.
const DefaultFriction = 0.6

// Frictioner is a block that has a friction value different than the default value of 0.6. The Friction
// method of the block returns the different friction value.
type Frictioner interface {
	Friction() float32
}
