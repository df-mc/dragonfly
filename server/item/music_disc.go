package item

import (
	"github.com/df-mc/dragonfly/server/world"
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

// AllMusicDiscs returns all 15 music disc items.
func AllMusicDiscs() []world.Item {
	m := make([]world.Item, 0, 15)
	for _, c := range sound.MusicDiscs() {
		m = append(m, MusicDisc{DiscType: c})
	}
	return m
}

// EncodeItem ...
func (m MusicDisc) EncodeItem() (name string, meta int16) {
	return "minecraft:music_disc_" + m.DiscType.String(), 0
}
