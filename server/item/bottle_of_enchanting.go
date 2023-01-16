package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// BottleOfEnchanting is a bottle that releases experience orbs when thrown.
type BottleOfEnchanting struct{}

// Use ...
func (b BottleOfEnchanting) Use(w *world.World, user User, ctx *UseContext) bool {
	create := w.EntityRegistry().Config().BottleOfEnchanting
	w.AddEntity(create(eyePosition(user), user.Rotation().Vec3().Mul(0.7), user))
	w.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (b BottleOfEnchanting) EncodeItem() (name string, meta int16) {
	return "minecraft:experience_bottle", 0
}
