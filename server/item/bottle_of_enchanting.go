package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// BottleOfEnchanting is a bottle that releases experience orbs when thrown.
type BottleOfEnchanting struct{}

// Dispense launches the bottle from a dispenser.
func (b BottleOfEnchanting) Dispense(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext) DispenseResult {
	create := tx.World().EntityRegistry().Config().BottleOfEnchanting
	if create == nil {
		return DispenseFailure
	}
	return dispenseProjectile(pos, face, tx, ctx, sound.ItemThrow{}, func(opts world.EntitySpawnOpts) *world.EntityHandle {
		return create(opts, nil)
	})
}

// Use ...
func (b BottleOfEnchanting) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().BottleOfEnchanting
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: throwableOffset(user.Rotation()).Vec3().Mul(0.6)}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (b BottleOfEnchanting) EncodeItem() (name string, meta int16) {
	return "minecraft:experience_bottle", 0
}
