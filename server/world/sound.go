package world

import (
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl64"
)

// Sound represents a sound that may be added to the world. When done, viewers of the world may be able to
// hear the sound.
type Sound interface {
	// Play plays the sound. This function may play other sounds too. It is always called when World.PlaySound
	// is called with the sound.
	Play(w *World, pos mgl64.Vec3)
}

// SoundHandles hands out the handles of sounds that are changed after they started playing, such as a music
// disc that is stopped. A server shares one between its worlds so that a handle is never used twice.
type SoundHandles struct {
	last atomic.Uint64
}

// Next returns a new sound handle. The first handle returned is 1.
func (h *SoundHandles) Next() uint64 {
	return h.last.Add(1)
}

// NextSoundHandle returns a new handle for a sound played in the World.
func (w *World) NextSoundHandle() uint64 {
	return w.conf.SoundHandles.Next()
}
