package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"time"
	_ "unsafe"
)

// Crossbow is a ranged weapon similar to a bow that uses arrows or fireworks as ammunition.
type Crossbow struct {
	// Item is the item the crossbow is charged with.
	Item Stack
}

// Charge starts the charging process and checks if the charge duration meets the required duration.
func (c Crossbow) Charge(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration) bool {
	if !c.Item.Empty() {
		return false
	}

	creative := releaser.GameMode().CreativeInventory()
	held, left := releaser.HeldItems()

	var projectileItem Stack
	if !left.Empty() {
		_, isFirework := left.Item().(Firework)
		_, isArrow := left.Item().(Arrow)
		if isFirework || isArrow {
			projectileItem = left
		}
	}

	if projectileItem.Empty() {
		var ok bool
		projectileItem, ok = ctx.FirstFunc(func(stack Stack) bool {
			_, isArrow := stack.Item().(Arrow)
			return isArrow
		})

		if !ok && !creative {
			return false
		}

		if projectileItem.Empty() {
			projectileItem = NewStack(Arrow{}, 1)
		}
	}

	c.Item = projectileItem.Grow(-projectileItem.Count() + 1)
	if !creative {
		ctx.Consume(c.Item)
	}

	crossbow := held.WithItem(c)
	releaser.SetHeldItems(crossbow, left)
	return true
}

// ReleaseCharge checks if the item is fully charged and, if so, releases it.
func (c Crossbow) ReleaseCharge(releaser Releaser, tx *world.Tx, ctx *UseContext) bool {
	if c.Item.Empty() {
		return false
	}

	creative := releaser.GameMode().CreativeInventory()
	rot := releaser.Rotation().Neg()
	dirVec := releaser.Rotation().Vec3().Normalize()

	if firework, isFirework := c.Item.Item().(Firework); isFirework {
		createFirework := tx.World().EntityRegistry().Config().Firework
		fireworkEntity := createFirework(world.EntitySpawnOpts{
			Position: torsoPosition(releaser),
			Velocity: dirVec.Mul(0.1),
			Rotation: rot,
		}, firework, releaser, 1.15, 0, false)
		tx.AddEntity(fireworkEntity)
		ctx.DamageItem(3)
	} else {
		createArrow := tx.World().EntityRegistry().Config().Arrow
		arrow := createArrow(world.EntitySpawnOpts{
			Position: torsoPosition(releaser),
			Velocity: dirVec.Mul(5.15),
			Rotation: rot,
		}, 9, releaser, false, false, !creative, 0, c.Item.Item().(Arrow).Tip)
		tx.AddEntity(arrow)
		ctx.DamageItem(1)
	}

	c.Item = Stack{}
	held, left := releaser.HeldItems()
	crossbow := held.WithItem(c)
	releaser.SetHeldItems(crossbow, left)
	tx.PlaySound(releaser.Position(), sound.CrossbowShoot{})
	return true
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
		return map[string]any{
			"chargedItem": writeItem(c.Item, true),
		}
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
