package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Egg is an item that can be used to craft food items, or as a throwable entity to spawn chicks.
type Egg struct{}

// MaxCount ...
func (e Egg) MaxCount() int {
	return 16
}

// Use ...
func (e Egg) Use(w *world.World, user User, ctx *UseContext) bool {
	create := w.EntityRegistry().Config().Egg
	w.AddEntity(create(eyePosition(user), user.Rotation().Vec3().Mul(1.5), user))
	w.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (e Egg) EncodeItem() (name string, meta int16) {
	return "minecraft:egg", 0
}
