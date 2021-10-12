package sound

// Named is a sound from resource pack, You can use custom sound from resource packs.
type Named struct {
	// SoundName is the name of the sound to play.
	SoundName string
	// Volume is the relative volume of the sound to play. It will be less loud for the player if it is
	// farther away from the position of the sound.
	Volume float32
	// Pitch is the pitch of the sound to play. Some sounds completely ignore this field, whereas others use
	// it to specify the pitch as the field is intended.
	Pitch float32

	sound
}
