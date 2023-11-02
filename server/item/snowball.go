package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Snowball is a throwable combat item obtained through shovelling snow.
type Snowball struct{}

// MaxCount ...
func (s Snowball) MaxCount() int {
	return 16
}

// Use ...
func (s Snowball) Use(w *world.World, user User, ctx *UseContext) bool {
	create := w.EntityRegistry().Config().Snowball
	w.AddEntity(create(eyePosition(user), user.Rotation().Vec3().Mul(1.5), user))
	w.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (s Snowball) EncodeItem() (name string, meta int16) {
	return "minecraft:snowball", 0
}
