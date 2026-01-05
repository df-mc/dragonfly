package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// WindCharge is a consumable item that can be thrown to create a wind burst.
type WindCharge struct{}

// MaxCount ...
func (w WindCharge) MaxCount() int {
	return 64
}

// Use ...
func (w WindCharge) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().WindCharge
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: user.Rotation().Vec3().Mul(1.5)}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (w WindCharge) EncodeItem() (name string, meta int16) {
	return "minecraft:wind_charge", 0
}