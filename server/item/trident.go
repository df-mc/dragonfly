package item

import (
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Trident is a weapon that can be used to perform melee attacks, or be thrown as a projectile.
type Trident struct{}

// MaxCount ...
func (Trident) MaxCount() int {
	return 1
}

// AttackDamage ...
func (Trident) AttackDamage() float64 {
	return 9
}

// HandEquipped ...
func (Trident) HandEquipped() bool {
	return true
}

// EnchantmentValue ...
func (Trident) EnchantmentValue() int {
	return 1
}

// DurabilityInfo ...
func (Trident) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability:    250,
		BrokenItem:       simpleItem(Stack{}),
		AttackDurability: 1,
		BreakDurability:  2,
	}
}

// RepairableBy ...
func (Trident) RepairableBy(i Stack) bool {
	_, ok := i.Item().(Trident)
	return ok
}

// Release either throws the trident as a projectile or, if the trident is
// enchanted with riptide, launches the releaser in the direction it is facing.
func (Trident) Release(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration) {
	if duration.Milliseconds()/50 < 10 {
		// The trident must be charged for at least ten ticks before it can be released.
		return
	}
	held, left := releaser.HeldItems()

	riptide := 0
	for _, enchant := range held.Enchantments() {
		if _, ok := enchant.Type().(interface{ RiptideForce(int) float64 }); ok {
			riptide = enchant.Level()
		}
	}
	if riptide > 0 {
		if !touchingWaterOrRain(releaser, tx) {
			return
		}
		// The riptide launch motion is handled client-side.
		if s, ok := releaser.(interface{ StartSpinning() }); ok {
			s.StartSpinning()
		}
		ctx.DamageItem(1)

		tx.PlaySound(releaser.Position(), sound.TridentRiptide{Level: riptide})
		return
	}

	creative := releaser.GameMode().CreativeInventory()
	thrown := held.Grow(-held.Count() + 1)
	if !creative {
		dmg := 1
		for _, enchant := range held.Enchantments() {
			if u, ok := enchant.Type().(interface {
				Reduce(it world.Item, level, amount int) int
			}); ok {
				dmg = u.Reduce(held.Item(), enchant.Level(), dmg)
			}
		}
		thrown = thrown.Damage(dmg)
		releaser.SetHeldItems(Stack{}, left)
	}

	if thrown.Empty() {
		tx.PlaySound(releaser.Position(), sound.ItemBreak{})
		return
	}
	create := tx.World().EntityRegistry().Config().Trident
	opts := world.EntitySpawnOpts{
		Position: eyePosition(releaser),
		Velocity: releaser.Rotation().Vec3().Mul(2.5),
		Rotation: releaser.Rotation().Neg(),
	}
	tx.AddEntity(create(opts, world.TridentSpawnConfig{
		Damage:        8,
		Owner:         releaser,
		Item:          thrown,
		DisablePickup: creative,
	}))
	tx.PlaySound(releaser.Position(), sound.TridentThrow{})
}

// touchingWaterOrRain checks if the world.Entity passed is currently either
// standing in water or exposed to rain.
func touchingWaterOrRain(e world.Entity, tx *world.Tx) bool {
	pos := cube.PosFromVec3(e.Position())
	if tx.RainingAt(pos) {
		return true
	}
	if l, ok := tx.Liquid(pos); ok && l.LiquidType() == "water" {
		return true
	}
	l, ok := tx.Liquid(cube.PosFromVec3(eyePosition(e)))
	return ok && l.LiquidType() == "water"
}

// Requirements returns the required items to release this item.
func (Trident) Requirements() []Stack {
	return []Stack{}
}

// EncodeItem ...
func (Trident) EncodeItem() (name string, meta int16) {
	return "minecraft:trident", 0
}
