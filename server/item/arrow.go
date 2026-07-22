package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Arrow is used as ammunition for bows, crossbows, and dispensers. Arrows can be modified to
// imbue status effects on players and mobs.
type Arrow struct {
	// Tip is the potion effect that is tipped on the arrow.
	Tip potion.Potion
}

// Dispense launches the arrow from a dispenser.
func (a Arrow) Dispense(pos cube.Pos, face cube.Face, tx *world.Tx, ctx *DispenseContext) DispenseResult {
	create := tx.World().EntityRegistry().Config().Arrow
	if create == nil {
		return DispenseFailure
	}
	return dispenseProjectile(pos, face, tx, ctx, sound.BowShoot{}, func(opts world.EntitySpawnOpts) *world.EntityHandle {
		return create(opts, world.ArrowSpawnConfig{Damage: 2, ObtainArrowOnPickup: true, Tip: a.Tip})
	})
}

// EncodeItem ...
func (a Arrow) EncodeItem() (name string, meta int16) {
	if tip := a.Tip.Uint8(); tip > 4 {
		return "minecraft:arrow", int16(tip + 1)
	}
	return "minecraft:arrow", 0
}

// OffHand ...
func (Arrow) OffHand() bool {
	return true
}
