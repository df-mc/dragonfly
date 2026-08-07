package sound

// Custom is a sound identified by name. It may be used to play sounds defined
// by a resource pack.
type Custom struct {
	// Name is the identifier of the sound.
	Name string
	// Volume is the volume of the sound.
	Volume float64
	// Pitch is the pitch of the sound.
	Pitch float64

	sound
}
