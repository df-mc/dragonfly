package iteminternal

import (
	"strings"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Components returns all the components of the given custom item. If the item has no components, a nil map and false
// are returned.
func Components(it world.CustomItem) map[string]any {
	category := it.Category()
	identifier, _ := it.EncodeItem()
	name := strings.Split(identifier, ":")[1]

	builder := NewComponentBuilder(it.Name(), identifier, category)

	if x, ok := it.(item.Armour); ok {
		builder.AddComponent("minecraft:armor", map[string]any{
			"protection": int32(x.DefencePoints()),
		})

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
			"slot": slot,
		})
	}
	if x, ok := it.(item.Consumable); ok {
		food := map[string]any{
			"can_always_eat": x.AlwaysConsumable(),
		}
		if y, ok := it.(item.Food); ok {
			info := y.FoodInfo()
			food["nutrition"] = int32(info.Nutrition)
			food["saturation_modifier"] = float32(info.SaturationModifier)
			if info.UsingConvertsTo != "" {
				food["using_converts_to"] = info.UsingConvertsTo
			}
			if info.OnUseAction != 0 {
				food["on_use_action"] = int32(info.OnUseAction)
			}
			if info.CooldownTime != 0 {
				food["cooldown_time"] = int32(info.CooldownTime)
				food["cooldown_type"] = info.CooldownType
			}
			if len(info.Effects) != 0 {
				effects := make([]any, 0, len(info.Effects))
				for _, e := range info.Effects {
					m := map[string]any{
						"id":        int32(e.ID),
						"duration":  int32(e.Duration),
						"amplifier": int32(e.Amplifier),
						"chance":    float32(e.Chance),
					}
					if e.Name != "" {
						m["name"] = e.Name
					}
					effects = append(effects, m)
				}
				food["effects"] = effects
			}
			if len(info.RemoveEffects) != 0 {
				removeEffects := make([]any, 0, len(info.RemoveEffects))
				for _, id := range info.RemoveEffects {
					removeEffects = append(removeEffects, int32(id))
				}
				food["remove_effects"] = removeEffects
			}
		}
		builder.AddComponent("minecraft:food", food)

		builder.AddProperty("use_duration", int32(x.ConsumeDuration().Seconds()*20))
		if y, ok := it.(item.Drinkable); ok && y.Drinkable() {
			builder.AddProperty("use_animation", int32(2))
		} else {
			builder.AddProperty("use_animation", int32(1))
		}
	}
	if x, ok := it.(item.Cooldown); ok {
		cooldown := map[string]any{
			"category": name,
			"duration": float32(x.Cooldown().Seconds()),
		}
		if y, ok := it.(item.CooldownTyped); ok {
			cooldown["type"] = y.CooldownType()
		}
		builder.AddComponent("minecraft:cooldown", cooldown)
	}
	if x, ok := it.(item.Durable); ok {
		builder.AddComponent("minecraft:durability", map[string]any{
			"max_durability": int32(x.DurabilityInfo().MaxDurability),
		})
	}
	if x, ok := it.(item.MaxCounter); ok {
		builder.AddProperty("max_stack_size", int32(x.MaxCount()))
	}
	if x, ok := it.(item.OffHand); ok {
		builder.AddProperty("allow_off_hand", x.OffHand())
	}
	if x, ok := it.(item.StackedByData); ok {
		builder.AddProperty("stacked_by_data", x.StackedByData())
	}
	if x, ok := it.(item.Throwable); ok {
		// The data in minecraft:projectile is only used by vanilla server-side, but we must send at least an empty map
		// so the client will play the throwing animation.
		builder.AddComponent("minecraft:projectile", map[string]any{})
		builder.AddComponent("minecraft:throwable", map[string]any{
			"do_swing_animation": x.SwingAnimation(),
		})
	}
	if x, ok := it.(item.Glinted); ok {
		builder.AddProperty("foil", x.Glinted())
	}
	if x, ok := it.(item.HandEquipped); ok {
		builder.AddProperty("hand_equipped", x.HandEquipped())
	}
	if x, ok := it.(item.Weapon); ok {
		builder.AddComponent("minecraft:damage", map[string]any{
			"value": int32(x.AttackDamage()),
		})
	}
	if x, ok := it.(item.Fuel); ok {
		builder.AddComponent("minecraft:fuel", map[string]any{
			"duration": float32(x.FuelInfo().Duration.Seconds()),
		})
	}
	if x, ok := it.(item.FireResistant); ok {
		builder.AddComponent("minecraft:fire_resistant", map[string]any{
			"value": x.FireResistant(),
		})
	}
	if x, ok := it.(item.EnchantableData); ok {
		info := x.EnchantableData()
		builder.AddComponent("minecraft:enchantable", map[string]any{
			"slot":  info.Slot,
			"value": int32(info.Value),
		})
	}
	if x, ok := it.(item.RepairMaterials); ok {
		builder.AddComponent("minecraft:repairable", map[string]any{
			"repair_items": repairItems(x.RepairMaterials()),
		})
	}
	if x, ok := it.(item.Tagged); ok {
		tags := stringSlice(x.Tags())
		builder.AddComponent("item_tags", tags)
		builder.AddComponent("minecraft:tags", map[string]any{"tags": tags})
	}
	if x, ok := it.(item.Seed); ok {
		info := x.SeedInfo()
		builder.AddComponent("minecraft:seed", map[string]any{
			"crop_result":                info.CropResult,
			"plant_at":                   stringSlice(info.PlantAt),
			"plant_at_any_solid_surface": info.PlantAtAnySolidSurface,
			"plant_at_face":              info.PlantAtFace,
		})
	}
	if x, ok := it.(item.Storage); ok {
		info := x.StorageInfo()
		if info.NumViewableSlots != 0 {
			builder.AddComponent("minecraft:bundle_interaction", map[string]any{
				"num_viewable_slots": int32(info.NumViewableSlots),
			})
		}
		if info.MaxSlots != 0 || len(info.AllowedItems) != 0 || len(info.BannedItems) != 0 {
			builder.AddComponent("minecraft:storage_item", map[string]any{
				"allow_nested_storage_items": info.AllowNestedStorageItems,
				"allowed_items":              stringSlice(info.AllowedItems),
				"banned_items":               bannedItems(info.BannedItems),
				"max_slots":                  int32(info.MaxSlots),
			})
		}
		if info.MaxWeightLimit != 0 {
			builder.AddComponent("minecraft:storage_weight_limit", map[string]any{
				"max_weight_limit": int32(info.MaxWeightLimit),
			})
		}
		if info.WeightInStorageItem != 0 {
			builder.AddComponent("minecraft:storage_weight_modifier", map[string]any{
				"weight_in_storage_item": int32(info.WeightInStorageItem),
			})
		}
	}
	if x, ok := it.(item.UseModifiers); ok {
		info := x.UseModifiers()
		modifiers := map[string]any{
			"emit_vibrations":   info.EmitVibrations,
			"movement_modifier": float32(info.MovementModifier),
			"use_duration":      float32(info.UseDuration),
		}
		if info.StartSound != "" {
			modifiers["start_sound"] = info.StartSound
		}
		if info.StartUsing != "" {
			modifiers["start_using"] = info.StartUsing
		}
		builder.AddComponent("minecraft:use_modifiers", modifiers)
	}
	if x, ok := it.(item.SwingDuration); ok {
		builder.AddComponent("minecraft:swing_duration", map[string]any{
			"value": float32(x.SwingDuration()),
		})
	}
	if x, ok := it.(item.SwingSounds); ok {
		info := x.SwingSounds()
		builder.AddComponent("minecraft:swing_sounds", map[string]any{
			"attack_hit":  info.AttackHit,
			"attack_miss": info.AttackMiss,
		})
	}
	if x, ok := it.(item.KineticWeapon); ok {
		builder.AddComponent("minecraft:kinetic_weapon", map[string]any{
			"minecraft:kinetic_weapon": kineticWeaponData(x.KineticWeaponInfo()),
		})
	}
	if x, ok := it.(item.PiercingWeapon); ok {
		info := x.PiercingWeaponInfo()
		builder.AddComponent("minecraft:piercing_weapon", map[string]any{
			"creative_reach": rangeData(info.CreativeReach),
			"hitbox_margin":  float32(info.HitboxMargin),
			"reach":          rangeData(info.Reach),
		})
	}
	if x, ok := it.(item.Camera); ok {
		info := x.CameraInfo()
		builder.AddComponent("minecraft:camera", map[string]any{
			"black_bars_duration":     float32(info.BlackBarsDuration),
			"black_bars_screen_ratio": float32(info.BlackBarsScreenRatio),
			"picture_duration":        float32(info.PictureDuration),
			"shutter_duration":        float32(info.ShutterDuration),
			"shutter_screen_ratio":    float32(info.ShutterScreenRatio),
			"slide_away_duration":     float32(info.SlideAwayDuration),
		})
		builder.AddComponent("minecraft:block", "minecraft:camera")
		if info.UseDuration != 0 {
			builder.AddProperty("use_duration", int32(info.UseDuration))
		}
	}
	return builder.Construct()
}

// repairItems converts the repair materials of an item to the data required for the minecraft:repairable
// component.
func repairItems(items []item.RepairItem) []any {
	materials := make([]any, 0, len(items))
	for _, r := range items {
		var entry []any
		if r.Item != "" {
			entry = []any{map[string]any{"name": r.Item}}
		} else {
			entry = []any{map[string]any{"tags": r.Tag}}
		}
		materials = append(materials, map[string]any{
			"items":         entry,
			"repair_amount": r.RepairAmount,
		})
	}
	return materials
}

// bannedItems converts the banned item identifiers of a storage item to the data required for the
// minecraft:storage_item component.
func bannedItems(items []string) []any {
	banned := make([]any, 0, len(items))
	for _, b := range items {
		banned = append(banned, map[string]any{"name": b})
	}
	return banned
}

// rangeData converts a min/max range to the data required for the client.
func rangeData(r [2]float64) map[string]any {
	return map[string]any{
		"min": float32(r[0]),
		"max": float32(r[1]),
	}
}

// kineticWeaponData converts the kinetic weapon information of an item to the data required for the
// minecraft:kinetic_weapon component.
func kineticWeaponData(info item.KineticWeaponInfo) map[string]any {
	conditions := func(c item.WeaponConditions) map[string]any {
		return map[string]any{
			"max_duration":       int32(c.MaxDuration),
			"min_relative_speed": float32(c.MinRelativeSpeed),
			"min_speed":          float32(c.MinSpeed),
		}
	}
	return map[string]any{
		"creative_reach":       rangeData(info.CreativeReach),
		"damage_conditions":    conditions(info.DamageConditions),
		"damage_modifier":      float32(info.DamageModifier),
		"damage_multiplier":    float32(info.DamageMultiplier),
		"delay":                int32(info.Delay),
		"dismount_conditions":  conditions(info.DismountConditions),
		"hitbox_margin":        float32(info.HitboxMargin),
		"knockback_conditions": conditions(info.KnockbackConditions),
		"reach":                rangeData(info.Reach),
	}
}

// stringSlice converts a slice of strings to a slice of any.
func stringSlice(x []string) []any {
	s := make([]any, len(x))
	for i, v := range x {
		s[i] = v
	}
	return s
}
