package item

import (
	"time"
	_ "unsafe"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Crossbow is a ranged weapon similar to a bow that uses arrows or fireworks
// as ammunition.
type Crossbow struct {
	// Item is the item the crossbow is charged with.
	Item Stack
}

// Charge starts the charging process and checks if the charge duration meets
// the required duration.
func (c Crossbow) Charge(releaser Releaser, _ *world.Tx, ctx *UseContext, duration time.Duration) bool {
	if !c.Item.Empty() {
		return false
	}

	creative := releaser.GameMode().CreativeInventory()
	held, left := releaser.HeldItems()

	if chargeDuration, _ := c.chargeDuration(held); duration < chargeDuration {
		return false
	}
	projectileItem, ok := c.findProjectile(releaser, ctx)
	if !ok {
		return false
	}
	c.Item = projectileItem.Grow(-projectileItem.Count() + 1)
	if !creative {
		ctx.Consume(c.Item)
	}

	releaser.SetHeldItems(held.WithItem(c), left)
	return true
}

// ContinueCharge ...
func (c Crossbow) ContinueCharge(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration) {
	if !c.Item.Empty() {
		return
	}

	held, _ := releaser.HeldItems()
	if _, ok := c.findProjectile(releaser, ctx); !ok {
		return
	}

	chargeDuration, qcLevel := c.chargeDuration(held)
	if duration.Seconds() <= 0.1 {
		tx.PlaySound(releaser.Position(), sound.CrossbowLoad{Stage: sound.CrossbowLoadingStart, QuickCharge: qcLevel > 0})
	}

	// Base reload time is 25 ticks; each Quick Charge level reduces by 5 ticks
	multiplier := 25.0 / float64(25-(5*qcLevel))

	// Adjust ticks based on the multiplier
	adjustedTicks := int(float64(duration.Milliseconds()) / (50 / multiplier))

	// Play sound after every 16 ticks (adjusted by Quick Charge)
	if adjustedTicks%16 == 0 {
		tx.PlaySound(releaser.Position(), sound.CrossbowLoad{Stage: sound.CrossbowLoadingMiddle, QuickCharge: qcLevel > 0})
	}

	if progress := float64(duration) / float64(chargeDuration); progress >= 1 {
		tx.PlaySound(releaser.Position(), sound.CrossbowLoad{Stage: sound.CrossbowLoadingEnd, QuickCharge: qcLevel > 0})
	}
}

// chargeDuration calculates the duration required to charge the crossbow and
// the quick charge enchantment level, if any.
func (c Crossbow) chargeDuration(s Stack) (dur time.Duration, quickChargeLvl int) {
	dur, lvl := time.Duration(1.25*float64(time.Second)), 0
	for _, enchant := range s.Enchantments() {
		if q, ok := enchant.Type().(interface{ ChargeDuration(int) time.Duration }); ok {
			dur = min(dur, q.ChargeDuration(enchant.Level()))
			lvl = enchant.Level()
		}
	}
	return dur, lvl
}

// findProjectile looks through the inventory of a Releaser to find a projectile
// to insert into the crossbow. It first checks the left hand for fireworks or
// arrows, and searches the rest of the inventory for arrows if no valid
// projectile was in the left hand. False is returned if no valid projectile was
// anywhere in the inventory.
func (c Crossbow) findProjectile(r Releaser, ctx *UseContext) (Stack, bool) {
	_, left := r.HeldItems()
	_, isFirework := left.Item().(Firework)
	_, isArrow := left.Item().(Arrow)
	if isFirework || isArrow {
		return left, true
	}
	if res, ok := ctx.FirstFunc(func(stack Stack) bool {
		_, ok := stack.Item().(Arrow)
		return ok
	}); ok {
		return res, true
	}
	if r.GameMode().CreativeInventory() {
		// No projectiles in inventory but the player is in creative mode, so
		// return an arrow anyway.
		return NewStack(Arrow{}, 1), true
	}
	return Stack{}, false
}

// ReleaseCharge checks if the item is fully charged and, if so, releases it.
func (c Crossbow) ReleaseCharge(releaser Releaser, tx *world.Tx, ctx *UseContext) bool {
	if c.Item.Empty() {
		return false
	}

	held, _ := releaser.HeldItems()
	creative := releaser.GameMode().CreativeInventory()

	multishot := false
	for _, enchant := range held.Enchantments() {
		if _, ok := enchant.Type().(interface{ MultipleProjectiles() bool }); ok {
			multishot = true
			break
		}
	}

	c.shoot(releaser, tx, 0, !creative)
	if multishot {
		c.shoot(releaser, tx, -10, false)
		c.shoot(releaser, tx, 10, false)
	}
	c.applyDamage(ctx)

	c.Item = Stack{}
	held, left := releaser.HeldItems()
	crossbow := held.WithItem(c)
	releaser.SetHeldItems(crossbow, left)
	tx.PlaySound(releaser.Position(), sound.CrossbowShoot{})
	return true
}

// shoot fires the crossbow's loaded projectiles.
func (c Crossbow) shoot(releaser Releaser, tx *world.Tx, offsetAngle float64, canObtainPickup bool) {
	rot := releaser.Rotation()
	dirVec := cube.Rotation{rot[0] + offsetAngle, rot[1]}.Vec3()

	if firework, ok := c.Item.Item().(Firework); ok {
		createFirework := tx.World().EntityRegistry().Config().Firework
		projectile := createFirework(world.EntitySpawnOpts{
			Position: torsoPosition(releaser),
			Velocity: dirVec.Mul(0.8),
			Rotation: rot.Neg(),
		}, firework, releaser, 1.0, 0, false)
		tx.AddEntity(projectile)
	} else {
		createArrow := tx.World().EntityRegistry().Config().Arrow
		arrow := createArrow(world.EntitySpawnOpts{
			Position: torsoPosition(releaser),
			Velocity: dirVec.Mul(5.15),
			Rotation: rot.Neg(),
		}, 9, releaser, false, false, canObtainPickup, 0, c.Item.Item().(Arrow).Tip)
		tx.AddEntity(arrow)
	}
}

// applyDamage applies damage on a UseContext based on the projectile loaded
// in the crossboww.
func (c Crossbow) applyDamage(ctx *UseContext) {
	if _, ok := c.Item.Item().(Firework); ok {
		ctx.DamageItem(3)
	} else {
		ctx.DamageItem(1)
	}
}

// MaxCount always returns 1.
func (Crossbow) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (Crossbow) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 464,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// FuelInfo ...
func (Crossbow) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// EnchantmentValue ...
func (Crossbow) EnchantmentValue() int {
	return 1
}

// EncodeItem ...
func (Crossbow) EncodeItem() (name string, meta int16) {
	return "minecraft:crossbow", 0
}

// DecodeNBT ...
func (c Crossbow) DecodeNBT(data map[string]any) any {
	c.Item = mapItem(data, "chargedItem")
	return c
}

// EncodeNBT ...
func (c Crossbow) EncodeNBT() map[string]any {
	if !c.Item.Empty() {
		return map[string]any{"chargedItem": writeItem(c.Item, true)}
	}
	return nil
}

// noinspection ALL
//
//go:linkname writeItem github.com/df-mc/dragonfly/server/internal/nbtconv.WriteItem
func writeItem(s Stack, disk bool) map[string]any

// noinspection ALL
//
//go:linkname mapItem github.com/df-mc/dragonfly/server/internal/nbtconv.MapItem
func mapItem(x map[string]any, k string) Stack
