package iteminternal

import (
	"fmt"
	"strings"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Components returns all the components of the given custom item. If the item has no components, a nil map and false
// are returned.
func Components(it world.CustomItem) (map[string]any, error) {
	category := it.Category()
	identifier, _ := it.EncodeItem()

	parts := strings.SplitN(identifier, ":", 1)
	if len(parts) < 2 {
		return nil, fmt.Errorf("identifier %s must contain namespace", identifier)
	}
	name := parts[1]

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
		builder.AddComponent("wearable", map[string]any{
			"slot":                  slot,
			"protection":            int32(x.DefencePoints()),
			"hides_player_location": x.HidesPlayerLocation(),
			"dispensable":           x.Dispensable(),
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
		builder.AddComponent("food", food)

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
		builder.AddComponent("cooldown", cooldown)
	}
	if x, ok := it.(item.Durable); ok {
		info := x.DurabilityInfo()
		damageChance := map[string]any{
			"min": int32(100),
			"max": int32(100),
		}
		if info.DamageChance != [2]int{} {
			damageChance["min"] = int32(info.DamageChance[0])
			damageChance["max"] = int32(info.DamageChance[1])
		}
		builder.AddComponent("durability", map[string]any{
			"max_durability": int32(info.MaxDurability),
			"damage_chance":  damageChance,
		})
	}
	if x, ok := it.(item.MaxCounter); ok {
		builder.AddProperty("max_stack_size", int32(x.MaxCount()))
	}
	if x, ok := it.(item.Icon); ok {
		textures := x.IconTextures()
		m := make(map[string]any, len(textures))
		for k, v := range textures {
			m[k] = v
		}
		builder.AddProperty("minecraft:icon", map[string]any{"textures": m})
	}
	if x, ok := it.(item.OffHand); ok {
		builder.AddProperty("allow_off_hand", x.OffHand())
	}
	if x, ok := it.(item.StackedByData); ok {
		builder.AddProperty("stacked_by_data", x.StackedByData())
	}
	if x, ok := it.(item.MiningSpeed); ok {
		builder.AddProperty("mining_speed", float32(x.MiningSpeed()))
	}
	if x, ok := it.(item.FrameCount); ok {
		builder.AddProperty("frame_count", int32(x.FrameCount()))
	}
	if x, ok := it.(item.CanDestroyInCreative); ok {
		builder.AddProperty("can_destroy_in_creative", x.CanDestroyInCreative())
	}
	if x, ok := it.(item.Projectile); ok {
		info := x.ProjectileInfo()
		builder.AddComponent("projectile", map[string]any{
			"minimum_critical_power": float32(info.MinimumCriticalPower),
			"projectile_entity":      info.ProjectileEntity,
		})
	} else if _, ok := it.(item.Throwable); ok {
		builder.AddComponent("projectile", map[string]any{})
	}
	if x, ok := it.(item.Throwable); ok {
		info := x.ThrowableInfo()
		throwable := map[string]any{
			"do_swing_animation": info.SwingAnimation,
		}
		if info.LaunchPowerScale != 0 {
			throwable["launch_power_scale"] = float32(info.LaunchPowerScale)
		}
		if info.MaxDrawDuration != 0 {
			throwable["max_draw_duration"] = float32(info.MaxDrawDuration)
		}
		if info.MaxLaunchPower != 0 {
			throwable["max_launch_power"] = float32(info.MaxLaunchPower)
		}
		if info.MinDrawDuration != 0 {
			throwable["min_draw_duration"] = float32(info.MinDrawDuration)
		}
		if info.ScalePowerByDrawDuration {
			throwable["scale_power_by_draw_duration"] = true
		}
		builder.AddComponent("throwable", throwable)
	}
	if x, ok := it.(item.Glinted); ok {
		builder.AddComponent("glint", map[string]any{
			"value": x.Glinted(),
		})
	}
	if x, ok := it.(item.HandEquipped); ok {
		builder.AddProperty("hand_equipped", x.HandEquipped())
	}
	if x, ok := it.(item.Weapon); ok {
		builder.AddComponent("damage", map[string]any{
			"value": x.AttackDamage(),
		})
	}
	if x, ok := it.(item.Fuel); ok {
		builder.AddComponent("fuel", map[string]any{
			"duration": float32(x.FuelInfo().Duration.Seconds()),
		})
	}
	if x, ok := it.(item.FireResistant); ok {
		builder.AddComponent("fire_resistant", map[string]any{
			"value": x.FireResistant(),
		})
	}
	if x, ok := it.(item.EnchantableData); ok {
		info := x.EnchantableData()
		builder.AddComponent("enchantable", map[string]any{
			"slot":  info.Slot,
			"value": info.Value,
		})
	}
	if x, ok := it.(item.RepairMaterials); ok {
		builder.AddComponent("repairable", map[string]any{
			"repair_items": repairItems(x.RepairMaterials()),
		})
	}
	if x, ok := it.(item.Tagged); ok {
		tags := stringSlice(x.Tags())
		builder.AddComponent("item_tags", tags)
		builder.AddComponent("tags", map[string]any{"tags": tags})
	}
	if x, ok := it.(item.Seed); ok {
		info := x.SeedInfo()
		builder.AddComponent("seed", map[string]any{
			"crop_result":                info.CropResult,
			"plant_at":                   stringSlice(info.PlantAt),
			"plant_at_any_solid_surface": info.PlantAtAnySolidSurface,
			"plant_at_face":              info.PlantAtFace,
		})
	}
	if x, ok := it.(item.Storage); ok {
		info := x.StorageInfo()
		if info.NumViewableSlots != 0 {
			if info.NumViewableSlots < 1 || info.NumViewableSlots > 64 {
				return nil, fmt.Errorf("NumViewableSlots %d out of range 1-64", info.NumViewableSlots)
			}
			builder.AddComponent("bundle_interaction", map[string]any{
				"num_viewable_slots": int32(info.NumViewableSlots),
			})
		}
		if info.MaxSlots != 0 || len(info.AllowedItems) != 0 || len(info.BannedItems) != 0 {
			builder.AddComponent("storage_item", map[string]any{
				"allow_nested_storage_items": info.AllowNestedStorageItems,
				"allowed_items":              stringSlice(info.AllowedItems),
				"banned_items":               bannedItems(info.BannedItems),
				"max_slots":                  int32(info.MaxSlots),
			})
		}
		if info.MaxWeightLimit != 0 {
			builder.AddComponent("storage_weight_limit", map[string]any{
				"max_weight_limit": int32(info.MaxWeightLimit),
			})
		}
		if info.WeightInStorageItem != 0 {
			builder.AddComponent("storage_weight_modifier", map[string]any{
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
		builder.AddComponent("use_modifiers", modifiers)
	}
	if x, ok := it.(item.SwingDuration); ok {
		builder.AddComponent("swing_duration", map[string]any{
			"value": float32(x.SwingDuration()),
		})
	}
	if x, ok := it.(item.SwingSounds); ok {
		info := x.SwingSounds()
		builder.AddComponent("swing_sounds", map[string]any{
			"attack_critical_hit": info.AttackCriticalHit,
			"attack_hit":          info.AttackHit,
			"attack_miss":         info.AttackMiss,
		})
	}
	if x, ok := it.(item.KineticWeapon); ok {
		builder.AddComponent("kinetic_weapon", map[string]any{
			"kinetic_weapon": kineticWeaponData(x.KineticWeaponInfo()),
		})
	}
	if x, ok := it.(item.PiercingWeapon); ok {
		info := x.PiercingWeaponInfo()
		builder.AddComponent("piercing_weapon", map[string]any{
			"creative_reach": rangeData(info.CreativeReach),
			"hitbox_margin":  float32(info.HitboxMargin),
			"reach":          rangeData(info.Reach),
		})
	}
	if x, ok := it.(item.Camera); ok {
		info := x.CameraInfo()
		builder.AddComponent("camera", map[string]any{
			"black_bars_duration":     float32(info.BlackBarsDuration),
			"black_bars_screen_ratio": float32(info.BlackBarsScreenRatio),
			"picture_duration":        float32(info.PictureDuration),
			"shutter_duration":        float32(info.ShutterDuration),
			"shutter_screen_ratio":    float32(info.ShutterScreenRatio),
			"slide_away_duration":     float32(info.SlideAwayDuration),
		})
		builder.AddComponent("block", "camera")
		if info.UseDuration != 0 {
			builder.AddProperty("use_duration", int32(info.UseDuration))
		}
	}
	if x, ok := it.(item.BlockPlacer); ok {
		info := x.BlockPlacerInfo()
		blockPlacer := map[string]any{
			"block":              info.Block,
			"replace_block_item": info.ReplaceBlockItem,
			"aligned_placement":  info.AlignedPlacement,
		}
		if len(info.UseOn) != 0 {
			blockPlacer["use_on"] = stringSlice(info.UseOn)
		}
		builder.AddComponent("block_placer", blockPlacer)
	}
	if x, ok := it.(item.Compostable); ok {
		builder.AddComponent("compostable", map[string]any{
			"composting_chance": int32(x.CompostChance() * 100),
		})
	}
	if x, ok := it.(item.DamageAbsorption); ok {
		builder.AddComponent("damage_absorption", map[string]any{
			"absorbable_causes": stringSlice(x.AbsorbableCauses()),
		})
	}
	if x, ok := it.(item.Digger); ok {
		info := x.DiggerInfo()
		speeds := make([]any, 0, len(info.DestroySpeeds))
		for _, ds := range info.DestroySpeeds {
			speeds = append(speeds, map[string]any{
				"block": ds.Block,
				"speed": float32(ds.Speed),
			})
		}
		builder.AddComponent("digger", map[string]any{
			"destroy_speeds": speeds,
			"use_efficiency": info.UseEfficiency,
		})
	}
	if x, ok := it.(item.DurabilitySensor); ok {
		info := x.DurabilitySensorInfo()
		sensor := map[string]any{}
		if info.SoundEvent != "" {
			sensor["sound_event"] = info.SoundEvent
		}
		if len(info.DurabilityThresholds) != 0 {
			sensor["durability_thresholds"] = durabilityThresholds(info.DurabilityThresholds)
		}
		builder.AddComponent("durability_sensor", sensor)
	}
	if x, ok := it.(item.Dyeable); ok {
		c := x.DefaultColor()
		builder.AddComponent("dyeable", map[string]any{
			"default_color": []any{int32(c[0]), int32(c[1]), int32(c[2])},
		})
	}
	if x, ok := it.(item.EntityPlacer); ok {
		info := x.EntityPlacerInfo()
		entityPlacer := map[string]any{
			"entity": info.Entity,
		}
		if len(info.UseOn) != 0 {
			entityPlacer["use_on"] = stringSlice(info.UseOn)
		}
		if len(info.DispenseOn) != 0 {
			entityPlacer["dispense_on"] = stringSlice(info.DispenseOn)
		}
		builder.AddComponent("entity_placer", entityPlacer)
	}
	if x, ok := it.(item.HoverTextColor); ok {
		builder.AddComponent("hover_text_color", map[string]any{
			"value": x.HoverTextColor(),
		})
	}
	if x, ok := it.(item.InteractButton); ok {
		builder.AddComponent("interact_button", map[string]any{
			"value": x.InteractButton(),
		})
	}
	if x, ok := it.(item.LiquidClipped); ok {
		builder.AddComponent("liquid_clipped", map[string]any{
			"value": x.LiquidClipped(),
		})
	}
	if x, ok := it.(item.Rarity); ok {
		builder.AddComponent("rarity", map[string]any{
			"value": x.Rarity(),
		})
	}
	if x, ok := it.(item.Record); ok {
		info := x.RecordInfo()
		builder.AddComponent("record", map[string]any{
			"comparator_signal": int32(info.ComparatorSignal),
			"duration":          float32(info.Duration),
			"sound_event":       info.SoundEvent,
		})
	}
	if x, ok := it.(item.Shooter); ok {
		info := x.ShooterInfo()
		builder.AddComponent("shooter", map[string]any{
			"ammunition":                   ammunition(info.Ammunition),
			"charge_on_draw":               info.ChargeOnDraw,
			"max_draw_duration":            float32(info.MaxDrawDuration),
			"scale_power_by_draw_duration": info.ScalePowerByDrawDuration,
		})
	}
	if x, ok := it.(item.ShouldDespawn); ok {
		builder.AddComponent("should_despawn", map[string]any{
			"value": x.ShouldDespawn(),
		})
	}
	return builder.Construct(), nil
}

// ammunition converts a slice of ammunition to the data required for the shooter component.
func ammunition(items []item.Ammunition) []any {
	ammo := make([]any, 0, len(items))
	for _, a := range items {
		ammo = append(ammo, map[string]any{
			"item":             a.Item,
			"search_inventory": a.SearchInventory,
			"use_in_creative":  a.UseInCreative,
			"use_offhand":      a.UseOffHand,
		})
	}
	return ammo
}

// repairItems converts the repair materials of an item to the data required for the repairable
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
// storage_item component.
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
// kinetic_weapon component.
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

// durabilityThresholds converts a slice of durability thresholds to the data required for the
// durability_sensor component.
func durabilityThresholds(thresholds []item.DurabilityThreshold) []any {
	t := make([]any, 0, len(thresholds))
	for _, th := range thresholds {
		m := map[string]any{
			"durability": int32(th.Durability),
		}
		if th.ParticleType != "" {
			m["particle_type"] = th.ParticleType
		}
		if th.SoundEvent != "" {
			m["sound_event"] = th.SoundEvent
		}
		t = append(t, m)
	}
	return t
}
