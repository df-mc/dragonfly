package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// LingeringPotion is a variant of a splash potion that can be thrown to leave clouds with status effects that linger on
// the ground in an area.
type LingeringPotion struct {
	// Type is the type of lingering potion.
	Type potion.Potion
}

// MaxCount ...
func (l LingeringPotion) MaxCount() int {
	return 1
}

// Use ...
func (l LingeringPotion) Use(w *world.World, user User, ctx *UseContext) bool {
	create := w.EntityRegistry().Config().LingeringPotion
	w.AddEntity(create(eyePosition(user), user.Rotation().Vec3().Mul(0.5), l.Type, user))
	w.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (l LingeringPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:lingering_potion", int16(l.Type.Uint8())
}
