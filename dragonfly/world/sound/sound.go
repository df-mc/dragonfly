package sound

// Sound represents a sound that may be added to the world. When done, viewers of the world may be able to
// hear the sound.
type Sound interface {
	__()
}

type sound struct{}

func (sound) __() {}
