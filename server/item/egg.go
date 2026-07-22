package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Egg is an item that can be used to craft food items, or as a throwable entity to spawn chicks.
type Egg struct{}

// Dispense launches the egg from a dispenser.
func (e Egg) Dispense(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext) DispenseResult {
	create := tx.World().EntityRegistry().Config().Egg
	if create == nil {
		return DispenseFailure
	}
	return dispenseProjectile(pos, face, tx, ctx, sound.ItemThrow{}, func(opts world.EntitySpawnOpts) *world.EntityHandle {
		return create(opts, nil)
	})
}

// MaxCount ...
func (e Egg) MaxCount() int {
	return 16
}

// Use ...
func (e Egg) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().Egg
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: user.Rotation().Vec3().Mul(1.5)}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (e Egg) EncodeItem() (name string, meta int16) {
	return "minecraft:egg", 0
}
