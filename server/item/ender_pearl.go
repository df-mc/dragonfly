package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"time"
)

// EnderPearl is a smooth, greenish-blue item used to teleport and to make an eye of ender.
type EnderPearl struct{}

func (e EnderPearl) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().EnderPearl
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: user.Rotation().Vec3().Mul(1.5)}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

func (EnderPearl) Cooldown() time.Duration {
	return time.Second
}

func (EnderPearl) MaxCount() int {
	return 16
}

func (EnderPearl) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_pearl", 0
}
