package item_internal

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/armour"
	"github.com/df-mc/dragonfly/server/world"
	"strings"
)

// ComponentsFromItem returns all the components of the given custom item. If the item has no components, a nil map and
// false are returned.
func ComponentsFromItem(it world.CustomItem) (map[string]interface{}, bool) {
	category := it.Category()
	identifier, _ := it.EncodeItem()
	name := strings.Split(identifier, ":")[1]

	itemProperties := map[string]interface{}{
		"minecraft:icon": map[string]interface{}{
			"texture": name,
		},
		"creative_group":    category.String(),
		"creative_category": int32(category.Uint8()),
		"max_stack_size":    int32(64),
	}
	components := map[string]interface{}{
		"item_properties": itemProperties,
		"minecraft:display_name": map[string]interface{}{
			"value": it.Name(),
		},
	}

	if x, ok := it.(armour.Armour); ok {
		components["minecraft:armor"] = map[string]interface{}{
			"protection": int32(x.DefencePoints()),
		}
		components["minecraft:knockback_resistance"] = map[string]interface{}{
			"protection": float32(x.KnockBackResistance()),
		}

		var slot int32
		switch it.(type) {
		case armour.Helmet:
			slot = 2
		case armour.Chestplate:
			slot = 3
		case armour.Leggings:
			slot = 4
		case armour.Boots:
			slot = 5
		}
		components["minecraft:wearable"] = map[string]interface{}{"slot": slot}
	}
	if x, ok := it.(item.Consumable); ok {
		itemProperties["use_duration"] = int32(x.ConsumeDuration().Seconds() * 20)

		if y, ok := it.(item.Drinkable); ok && y.Drinkable() {
			itemProperties["use_animation"] = int32(2)
		} else {
			itemProperties["use_animation"] = int32(1)
			// The data in minecraft:food is only used by vanilla server-side, but we must send at least an empty map so
			// the client will play the eating animation.
			components["minecraft:food"] = map[string]interface{}{}
		}
	}
	if x, ok := it.(item.Cooldown); ok {
		components["minecraft:cooldown"] = map[string]interface{}{
			"category": name,
			"duration": float32(x.Cooldown().Seconds()),
		}
	}
	if x, ok := it.(item.Durable); ok {
		components["minecraft:durability"] = map[string]interface{}{
			"max_durability": int32(x.DurabilityInfo().MaxDurability),
		}
	}
	if x, ok := it.(item.MaxCounter); ok {
		itemProperties["max_stack_size"] = int32(x.MaxCount())
	}
	if x, ok := it.(item.OffHand); ok {
		itemProperties["allow_off_hand"] = x.OffHand()
	}
	if x, ok := it.(item.Throwable); ok {
		// The data in minecraft:projectile is only used by vanilla server-side, but we must send at least an empty map
		// so the client will play the throwing animation.
		components["minecraft:projectile"] = map[string]interface{}{}
		components["minecraft:throwable"] = map[string]interface{}{
			"do_swing_animation": x.SwingAnimation(),
		}
	}

	// If an item has no new components or properties then it should not be considered a component-based item.
	if len(components) == 2 && len(itemProperties) == 4 && itemProperties["max_stack_size"] == int32(64) {
		return nil, false
	}
	return map[string]interface{}{"components": components}, true
}
