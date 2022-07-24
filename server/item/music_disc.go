package item

import (
	"github.com/df-mc/dragonfly/server/world/sound"
)

// MusicDisc is an item that can be played in jukeboxes.
type MusicDisc struct {
	// DiscType is the disc type of the music disc.
	DiscType sound.DiscType
}

// MaxCount always returns 1.
func (MusicDisc) MaxCount() int {
	return 1
}

// EncodeItem ...
func (m MusicDisc) EncodeItem() (name string, meta int16) {
	return "minecraft:music_disc_" + m.DiscType.String(), 0
}
