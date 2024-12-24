package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// BottleOfEnchanting is a bottle that releases experience orbs when thrown.
type BottleOfEnchanting struct{}

// Use ...
func (b BottleOfEnchanting) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().BottleOfEnchanting
	r := user.Rotation()
	r[1] = throwableOffset(r[1])
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: user.Rotation().Vec3().Mul(0.7)}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (b BottleOfEnchanting) EncodeItem() (name string, meta int16) {
	return "minecraft:experience_bottle", 0
}
