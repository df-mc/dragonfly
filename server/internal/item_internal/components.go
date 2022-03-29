package item_internal

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"strings"
)

// ComponentsFromItem returns all the components of the given custom item. If the item has no components, a nil map and
// false are returned.
func ComponentsFromItem(it world.CustomItem) (map[string]interface{}, bool) {
	category := it.Category()
	identifier, _ := it.EncodeItem()
	name := strings.Split(identifier, ":")[1]

	builder := NewComponentBuilder(it.Name(), identifier, category)

	if x, ok := it.(item.Armour); ok {
		builder.AddComponent("minecraft:armor", map[string]interface{}{
			"protection": int32(x.DefencePoints()),
		})
		builder.AddComponent("minecraft:knockback_resistance", map[string]interface{}{
			"protection": float32(x.KnockBackResistance()),
		})

		var slot int32
		switch it.(type) {
		case item.HelmetType:
			slot = 2
		case item.ChestplateType:
			slot = 3
		case item.LeggingsType:
			slot = 4
		case item.BootsType:
			slot = 5
		}
		builder.AddComponent("minecraft:wearable", map[string]interface{}{
			"slot": slot,
		})
	}
	if x, ok := it.(item.Consumable); ok {
		builder.AddItemProperty("use_duration", int32(x.ConsumeDuration().Seconds()*20))

		if y, ok := it.(item.Drinkable); ok && y.Drinkable() {
			builder.AddItemProperty("use_animation", int32(2))
		} else {
			builder.AddItemProperty("use_animation", int32(1))
			// The data in minecraft:food is only used by vanilla server-side, but we must send at least an empty map so
			// the client will play the eating animation.
			builder.AddComponent("minecraft:food", map[string]interface{}{})
		}
	}
	if x, ok := it.(item.Cooldown); ok {
		builder.AddComponent("minecraft:cooldown", map[string]interface{}{
			"category": name,
			"duration": float32(x.Cooldown().Seconds()),
		})
	}
	if x, ok := it.(item.Durable); ok {
		builder.AddComponent("minecraft:durability", map[string]interface{}{
			"max_durability": int32(x.DurabilityInfo().MaxDurability),
		})
	}
	if x, ok := it.(item.MaxCounter); ok {
		builder.AddItemProperty("max_stack_size", int32(x.MaxCount()))
	}
	if x, ok := it.(item.OffHand); ok {
		builder.AddItemProperty("allow_off_hand", x.OffHand())
	}
	if x, ok := it.(item.Throwable); ok {
		// The data in minecraft:projectile is only used by vanilla server-side, but we must send at least an empty map
		// so the client will play the throwing animation.
		builder.AddComponent("minecraft:projectile", map[string]interface{}{})
		builder.AddComponent("minecraft:throwable", map[string]interface{}{
			"do_swing_animation": x.SwingAnimation(),
		})
	}

	// If an item has no new components or properties then it should not be considered a component-based item.
	if builder.Empty() {
		return nil, false
	}
	return builder.Construct(), true
}
