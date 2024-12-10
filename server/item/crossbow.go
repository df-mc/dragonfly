package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"time"
	_ "unsafe"
)

type Crossbow struct {
	Item Stack
}

// Charge starts the charging process and prints the intended duration.
func (c Crossbow) Charge(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration) {
	creative := releaser.GameMode().CreativeInventory()
	held, left := releaser.HeldItems()

	chargeDuration := 1250 * time.Millisecond
	for _, enchant := range held.Enchantments() {
		if q, ok := enchant.Type().(interface{ ChargeDuration(int) time.Duration }); ok {
			chargeDuration = q.ChargeDuration(enchant.Level())
		}
	}

	if duration >= chargeDuration {
		var projectileItem Stack
		if !left.Empty() {
			if _, isFirework := left.Item().(Firework); isFirework {
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
				return
			}

			if projectileItem.Empty() {
				projectileItem = NewStack(Arrow{}, 1)
			}
		}

		c.Item = projectileItem.Grow(-projectileItem.Count() + 1)

		if !creative {
			ctx.Consume(c.Item)
		}

		crossbow := newCrossbowWith(held, c)
		releaser.SetHeldItems(crossbow, left)
	}
}

// Release handles the firing of the crossbow after charging.
func (c Crossbow) Release(releaser Releaser, tx *world.Tx, ctx *UseContext) {
	creative := releaser.GameMode().CreativeInventory()

	// Check if the projectile is a firework or an arrow.
	if _, isFirework := c.Item.Item().(Firework); isFirework {
		create := tx.World().EntityRegistry().Config().Firework
		firework := create(world.EntitySpawnOpts{
			Position: eyePosition(releaser),
			Velocity: releaser.Rotation().Vec3().Normalize().Mul(0.32),
			Rotation: releaser.Rotation(),
		}, c.Item.Item(), releaser, false)
		tx.AddEntity(firework)
	} else {
		var tip potion.Potion
		if !c.Item.Empty() {
			// Arrow is empty if not found in the creative inventory.
			tip = c.Item.Item().(Arrow).Tip
		}

		create := tx.World().EntityRegistry().Config().Arrow
		damage, consume := 6.0, !creative
		arrow := create(world.EntitySpawnOpts{
			Position: eyePosition(releaser),
			Velocity: releaser.Rotation().Vec3().Normalize().Mul(3),
			Rotation: releaser.Rotation(),
		}, damage, releaser, true, false, !creative && consume, 0, tip)

		tx.AddEntity(arrow)
	}

	releaser.PlaySound(sound.BowShoot{})

	c.Item = Stack{}
	held, left := releaser.HeldItems()
	crossbow := newCrossbowWith(held, c)
	releaser.SetHeldItems(crossbow, left)
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
	return newFuelInfo(time.Second * 10)
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

// newCrossbowWith duplicates an item.Stack with the new item type given.
func newCrossbowWith(input Stack, item Crossbow) Stack {
	if _, ok := input.Item().(Crossbow); !ok {
		return Stack{}
	}
	outputStack := NewStack(item, input.Count()).
		Damage(input.MaxDurability() - input.Durability()).
		WithCustomName(input.CustomName()).
		WithLore(input.Lore()...).
		WithEnchantments(input.Enchantments()...).
		WithAnvilCost(input.AnvilCost())
	for k, v := range input.Values() {
		outputStack = outputStack.WithValue(k, v)
	}
	return outputStack
}

// noinspection ALL
//
//go:linkname writeItem github.com/df-mc/dragonfly/server/internal/nbtconv.WriteItem
func writeItem(s Stack, disk bool) map[string]any

// noinspection ALL
//
//go:linkname mapItem github.com/df-mc/dragonfly/server/internal/nbtconv.MapItem
func mapItem(x map[string]any, k string) Stack
