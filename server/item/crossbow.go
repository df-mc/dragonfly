package item

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"time"
	_ "unsafe"
)

// Crossbow is a ranged weapon similar to a bow that uses arrows or fireworks as ammunition.
type Crossbow struct {
	// Item is the item the crossbow is charged with.
	Item Stack
}

// Charge starts the charging process and prints the intended duration.
func (c Crossbow) Charge(releaser Releaser, tx *world.Tx, ctx *UseContext, duration time.Duration) {
	if !c.Item.Empty() {
		return
	}

	creative := releaser.GameMode().CreativeInventory()
	held, left := releaser.HeldItems()

	chargeDuration := time.Duration(1.25 * float64(time.Second))
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
		return
	}
	return
}

func (c Crossbow) Release(releaser Releaser, tx *world.Tx, ctx *UseContext) bool {
	if c.Item.Empty() {
		return false
	}

	creative := releaser.GameMode().CreativeInventory()

	rot := releaser.Rotation()
	rot = cube.Rotation{-rot[0], -rot[1]}
	if rot[0] > 180 {
		rot[0] = 360 - rot[0]
	}

	dirVec := releaser.Rotation().Vec3().Normalize()
	if firework, isFirework := c.Item.Item().(Firework); isFirework {
		createFirework := tx.World().EntityRegistry().Config().Firework
		fireworkEntity := createFirework(world.EntitySpawnOpts{
			Position: torsoPosition(releaser),
			Velocity: dirVec.Mul(1.5),
			Rotation: rot,
		}, firework, releaser, false)
		tx.AddEntity(fireworkEntity)
		return true
	}

	createArrow := tx.World().EntityRegistry().Config().Arrow
	arrow := createArrow(world.EntitySpawnOpts{
		Position: torsoPosition(releaser),
		Velocity: dirVec.Mul(3.0),
		Rotation: rot,
	}, 9, releaser, true, false, !creative, 0, potion.Potion{})
	tx.AddEntity(arrow)

	c.Item = Stack{}
	held, left := releaser.HeldItems()
	crossbow := newCrossbowWith(held, c)
	releaser.SetHeldItems(crossbow, left)
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
