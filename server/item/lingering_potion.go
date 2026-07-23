package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
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

// Dispense launches the potion from a dispenser.
func (l LingeringPotion) Dispense(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext) DispenseResult {
	create := tx.World().EntityRegistry().Config().LingeringPotion
	if create == nil {
		return DispenseFailure
	}
	return dispenseProjectile(pos, face, tx, ctx, sound.ItemThrow{}, func(opts world.EntitySpawnOpts) *world.EntityHandle {
		return create(opts, l.Type, nil)
	})
}

// MaxCount ...
func (l LingeringPotion) MaxCount() int {
	return 1
}

// Use ...
func (l LingeringPotion) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().LingeringPotion
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: throwableOffset(user.Rotation()).Vec3().Mul(0.5)}
	tx.AddEntity(create(opts, l.Type, user.H()))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (l LingeringPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:lingering_potion", int16(l.Type.Uint8())
}
