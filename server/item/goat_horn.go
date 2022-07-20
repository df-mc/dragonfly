package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"time"
)

// GoatHorn is an item dropped by goats. It has eight variants, and each plays a unique sound when used which can be
// heard by players in a large radius.
type GoatHorn struct {
	nopReleasable

	// Type is the type of the goat horn, determining the sound it plays.
	Type sound.GeatHorn
}

// MaxCount ...
func (GoatHorn) MaxCount() int {
	return 1
}

// Cooldown ...
func (GoatHorn) Cooldown() time.Duration {
	return time.Second * 7
}

// Use ...
func (g GoatHorn) Use(w *world.World, u User, _ *UseContext) bool {
	w, pos := u.World(), u.Position()
	w.PlaySound(pos, sound.GoatHorn{Horn: g.Type})
	time.AfterFunc(time.Second, func() {
		// The goat horn is forcefully released by the server after a second. If the client released the item itself,
		// before a second, this shouldn't do anything.
		u.ReleaseItem()
	})
	return true
}

// EncodeItem ...
func (g GoatHorn) EncodeItem() (name string, meta int16) {
	return "minecraft:goat_horn", int16(g.Type.Uint8())
}
