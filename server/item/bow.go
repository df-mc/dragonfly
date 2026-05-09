package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math"
	"time"
)

// Bow is a ranged weapon that fires arrows.
type Bow struct{}

// MaxCount always returns 1.
func (Bow) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (Bow) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 385,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// FuelInfo ...
func (Bow) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 10)
}

// Release ...
func (Bow) Release(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration) {
	creative := releaser.GameMode().CreativeInventory()
	ticks := duration.Milliseconds() / 50
	if ticks < 3 {
		// The player must hold the bow for at least three ticks.
		return
	}

	d := float64(ticks) / 20
	force := math.Min((d*d+d*2)/3, 1)
	if force < 0.1 {
		// The force must be at least 0.1.
		return
	}

	arrow, ok := ctx.FirstFunc(func(stack Stack) bool {
		_, ok := stack.Item().(Arrow)
		return ok
	})
	if !ok && !creative {
		// No arrows in inventory and not in creative mode.
		return
	}

	var tip potion.Potion
	if !arrow.Empty() {
		// Arrow is empty if not found in the creative inventory.
		tip = arrow.Item().(Arrow).Tip
	}

	held, _ := releaser.HeldItems()
	damage, punchLevel, burnDuration, consume := 2.0, 0, time.Duration(0), !creative
	for _, enchant := range held.Enchantments() {
		if f, ok := enchant.Type().(interface{ BurnDuration() time.Duration }); ok {
			burnDuration = f.BurnDuration()
		}
		if _, ok := enchant.Type().(interface{ KnockBackMultiplier() float64 }); ok {
			punchLevel = enchant.Level()
		}
		if p, ok := enchant.Type().(interface{ PowerDamage(int) float64 }); ok {
			damage += p.PowerDamage(enchant.Level())
		}
		if i, ok := enchant.Type().(interface{ ConsumesArrows() bool }); ok && !i.ConsumesArrows() {
			consume = false
		}
	}

	create := tx.World().EntityRegistry().Config().Arrow
	opts := world.EntitySpawnOpts{Position: eyePosition(releaser), Velocity: releaser.Rotation().Vec3().Mul(force * 5), Rotation: releaser.Rotation().Neg()}
	projectile := tx.AddEntity(create(opts, damage, releaser, force >= 1, false, !creative && consume, punchLevel, 0, tip))
	if f, ok := projectile.(interface{ SetOnFire(duration time.Duration) }); ok {
		f.SetOnFire(burnDuration)
	}

	ctx.DamageItem(1)
	if consume {
		ctx.Consume(arrow.Grow(-arrow.Count() + 1))
	}

	tx.PlaySound(releaser.Position(), sound.BowShoot{})
}

// EnchantmentValue ...
func (Bow) EnchantmentValue() int {
	return 1
}

// Requirements returns the required items to release this item.
func (Bow) Requirements() []Stack {
	return []Stack{NewStack(Arrow{}, 1)}
}

// EncodeItem ...
func (Bow) EncodeItem() (name string, meta int16) {
	return "minecraft:bow", 0
}
