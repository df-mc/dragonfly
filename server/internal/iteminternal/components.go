package iteminternal

import (
	"fmt"
	"strings"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/component"
	"github.com/df-mc/dragonfly/server/world"
)

// Components returns all the components of the given custom item. If the item has no components, a nil map and false
// are returned.
func Components(it world.CustomItem) (map[string]any, error) {
	category := it.Category()
	identifier, _ := it.EncodeItem()

	_, name, ok := strings.Cut(identifier, ":")
	if !ok {
		return nil, fmt.Errorf("identifier %s must contain namespace", identifier)
	}

	// Check for new ComponentItem interface first
	if ci, ok := it.(component.ComponentItem); ok {
		comps, err := componentsFromComponentItem(ci)
		if err != nil {
			return nil, err
		}
		builder := NewComponentBuilder(it.Name(), identifier, category)
		// Add the components from ComponentItem
		for name, data := range comps {
			builder.AddComponent(name, data)
		}
		result := builder.Construct()
		components := result["components"].(map[string]any)
		if err := ValidateComponents(components); err != nil {
			return nil, fmt.Errorf("component validation failed for %s: %w", name, err)
		}
		return result, nil
	}

	builder := NewComponentBuilder(it.Name(), identifier, category)

	if x, ok := it.(item.Armour); ok {
		var slot string
		switch it.(type) {
		case item.HelmetType:
			slot = "slot.armor.head"
		case item.ChestplateType:
			slot = "slot.armor.chest"
		case item.LeggingsType:
			slot = "slot.armor.legs"
		case item.BootsType:
			slot = "slot.armor.feet"
		}
		builder.AddComponent("minecraft:wearable", map[string]any{
			"slot":                  slot,
			"protection":            int32(x.DefencePoints()),
			"hides_player_location": false,
			"dispensable":           false,
		})
	}
	if x, ok := it.(item.Consumable); ok {
		builder.AddProperty("use_duration", int32(x.ConsumeDuration().Seconds()*20))
		builder.AddComponent("minecraft:food", map[string]any{
			"can_always_eat": x.AlwaysConsumable(),
		})

		if y, ok := it.(item.Drinkable); ok && y.Drinkable() {
			builder.AddProperty("use_animation", int32(2))
		} else {
			builder.AddProperty("use_animation", int32(1))
		}
	}
	if x, ok := it.(item.Cooldown); ok {
		builder.AddComponent("minecraft:cooldown", map[string]any{
			"category": name,
			"duration": float32(x.Cooldown().Seconds()),
		})
	}
	if x, ok := it.(item.Durable); ok {
		builder.AddComponent("minecraft:durability", map[string]any{
			"max_durability": int32(x.DurabilityInfo().MaxDurability),
			"damage_chance": map[string]any{
				"min": int32(100),
				"max": int32(100),
			},
		})
	}
	if x, ok := it.(item.MaxCounter); ok {
		builder.AddProperty("max_stack_size", int32(x.MaxCount()))
	}
	if x, ok := it.(item.OffHand); ok {
		builder.AddProperty("allow_off_hand", x.OffHand())
	}
	if _, ok := it.(item.Throwable); ok {
		// The data in minecraft:projectile is only used by vanilla server-side, but we must send at least an empty map
		// so the client will play the throwing animation.
		builder.AddComponent("minecraft:projectile", map[string]any{})
	}
	if x, ok := it.(item.Throwable); ok {
		builder.AddComponent("minecraft:throwable", map[string]any{
			"do_swing_animation": x.SwingAnimation(),
		})
	}
	if x, ok := it.(item.Glinted); ok {
		builder.AddComponent("minecraft:glint", map[string]any{
			"value": x.Glinted(),
		})
	}
	if x, ok := it.(item.HandEquipped); ok {
		builder.AddProperty("hand_equipped", x.HandEquipped())
	}
	if x, ok := it.(item.Weapon); ok {
		builder.AddComponent("minecraft:damage", map[string]any{
			"value": x.AttackDamage(),
		})
	}
	if x, ok := it.(item.Fuel); ok {
		builder.AddComponent("minecraft:fuel", map[string]any{
			"duration": float32(x.FuelInfo().Duration.Seconds()),
		})
	}
	if x, ok := it.(item.Compostable); ok {
		builder.AddComponent("minecraft:compostable", map[string]any{
			"composting_chance": int32(x.CompostChance() * 100),
		})
	}
	result := builder.Construct()
	components := result["components"].(map[string]any)
	if err := ValidateComponents(components); err != nil {
		return nil, fmt.Errorf("component validation failed for %s: %w", name, err)
	}
	return result, nil
}

// componentsFromComponentItem converts a ComponentItem to a components map.
func componentsFromComponentItem(ci component.ComponentItem) (map[string]any, error) {
	components := make(map[string]any)
	for _, comp := range ci.ItemComponents() {
		data, err := comp.Encode()
		if err != nil {
			return nil, err
		}
		components[comp.ComponentName()] = data
	}
	return components, nil
}
