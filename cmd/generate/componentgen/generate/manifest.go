package generate

// Codec describes how a component field is encoded to its NBT representation.
type Codec string

const (
	// CodecDirect encodes the field's value directly under its NBT name.
	CodecDirect Codec = "direct"
	// CodecStringSlice encodes a []string as a []any.
	CodecStringSlice Codec = "string_slice"
	// CodecList3 encodes a [3]int32 as a []any.
	CodecList3 Codec = "list3"
	// CodecRange encodes a [2]float32 as a {"min","max"} compound.
	CodecRange Codec = "range"
	// CodecRangeInt encodes a [2]int32 as a {"min","max"} compound.
	CodecRangeInt Codec = "range_int"
	// CodecSlice encodes a []Nested struct as a []any of encoded compounds.
	CodecSlice Codec = "slice"
	// CodecScalarSlice encodes a []T where T is a scalar as a []any of values.
	CodecScalarSlice Codec = "scalar_slice"
	// CodecNested encodes a single Nested struct as a compound.
	CodecNested Codec = "nested"
	// CodecBanned encodes a []string as a []any of {"name"} compounds.
	CodecBanned Codec = "banned"
	// CodecNameOrTags encodes a []RepairItemEntry as {"name"} or {"tags"} compounds.
	CodecNameOrTags Codec = "name_or_tags"
)

// Field describes a single struct field of a component (or nested type).
type Field struct {
	// GoName is the exported Go field name (e.g. "MaxDurability").
	GoName string
	// NBTName is the wire key (e.g. "max_durability"). Empty when not emitted directly.
	NBTName string
	// Type is the Go field type (e.g. "int32", "[]string", "WeaponConditions").
	Type string
	// Codec selects the encode behaviour. Defaults to CodecDirect.
	Codec Codec
	// SliceElement, when Codec is slice/nested/name_or_tags, names the nested type
	// whose fields are declared in Fields.
	SliceElement string
	// Fields holds the nested type's fields for slice/nested/name_or_tags codecs.
	Fields []Field
	// OmitEmpty omits the key when the value is the zero value of its type.
	OmitEmpty bool
}

// Spec describes a single generated component type.
type Spec struct {
	// Struct is the exported Go struct name (e.g. "Food").
	Struct string
	// Name is the namespaced component name (e.g. "minecraft:food").
	Name string
	// Comment is the doc comment for the struct.
	Comment string
	// Fields lists the struct fields in declaration order.
	Fields []Field
	// ConstantGroup is a set of package-level constants emitted alongside the struct.
	ConstantGroup []string
	// ConstantType, when non-empty, names a new named string type for the
	// ConstantGroup constants (e.g. "SlotArmor"). The type declaration is emitted
	// before the constants, which then declare `Name ConstantType = "value"`.
	ConstantType string
}

// manifest lists every component that should be generated.
var manifest = []Spec{
	{
		Struct:  "Wearable",
		Name:    "minecraft:wearable",
		Comment: "Wearable represents the minecraft:wearable component.",
		Fields: []Field{
			{GoName: "Slot", NBTName: "slot", Type: "SlotArmor"},
			{GoName: "Protection", NBTName: "protection", Type: "int32"},
			{GoName: "HidesPlayerLocation", NBTName: "hides_player_location", Type: "bool"},
			{GoName: "Dispensable", NBTName: "dispensable", Type: "bool"},
		},
		ConstantType: "SlotArmor",
		ConstantGroup: []string{
			"SlotArmorHead  SlotArmor = \"slot.armor.head\"",
			"SlotArmorChest SlotArmor = \"slot.armor.chest\"",
			"SlotArmorLegs  SlotArmor = \"slot.armor.legs\"",
			"SlotArmorFeet  SlotArmor = \"slot.armor.feet\"",
		},
	},
	{
		Struct:  "Food",
		Name:    "minecraft:food",
		Comment: "Food represents the minecraft:food component.",
		Fields: []Field{
			{GoName: "Nutrition", NBTName: "nutrition", Type: "int32"},
			{GoName: "SaturationModifier", NBTName: "saturation_modifier", Type: "float32"},
			{GoName: "CanAlwaysEat", NBTName: "can_always_eat", Type: "bool"},
			{GoName: "UsingConvertsTo", NBTName: "using_converts_to", Type: "string", OmitEmpty: true},
			{GoName: "OnUseAction", NBTName: "on_use_action", Type: "int32", OmitEmpty: true},
			{GoName: "CooldownTime", NBTName: "cooldown_time", Type: "int32", OmitEmpty: true},
			{GoName: "CooldownType", NBTName: "cooldown_type", Type: "string", OmitEmpty: true},
			{GoName: "Effects", NBTName: "effects", Type: "[]FoodEffect", Codec: CodecSlice, SliceElement: "FoodEffect", OmitEmpty: true,
				Fields: []Field{
					{GoName: "ID", NBTName: "id", Type: "int32"},
					{GoName: "Duration", NBTName: "duration", Type: "int32"},
					{GoName: "Amplifier", NBTName: "amplifier", Type: "int32"},
					{GoName: "Chance", NBTName: "chance", Type: "float32"},
					{GoName: "Name", NBTName: "name", Type: "string", OmitEmpty: true},
				}},
			{GoName: "RemoveEffects", NBTName: "remove_effects", Type: "[]int32", Codec: CodecScalarSlice, OmitEmpty: true},
		},
	},
	{
		Struct:  "Durability",
		Name:    "minecraft:durability",
		Comment: "Durability represents the minecraft:durability component.",
		Fields: []Field{
			{GoName: "MaxDurability", NBTName: "max_durability", Type: "int32"},
			{GoName: "DamageChance", NBTName: "damage_chance", Type: "[2]int32", Codec: CodecRangeInt},
		},
	},
	{
		Struct:  "KineticWeapon",
		Name:    "minecraft:kinetic_weapon",
		Comment: "KineticWeapon represents the minecraft:kinetic_weapon component.",
		Fields: []Field{
			{GoName: "CreativeReach", NBTName: "creative_reach", Type: "[2]float32", Codec: CodecRange},
			{GoName: "DamageConditions", NBTName: "damage_conditions", Type: "WeaponConditions", Codec: CodecNested, SliceElement: "WeaponConditions",
				Fields: []Field{
					{GoName: "MaxDuration", NBTName: "max_duration", Type: "int32"},
					{GoName: "MinSpeed", NBTName: "min_speed", Type: "float32"},
					{GoName: "MinRelativeSpeed", NBTName: "min_relative_speed", Type: "float32"},
				}},
			{GoName: "DamageModifier", NBTName: "damage_modifier", Type: "float32"},
			{GoName: "DamageMultiplier", NBTName: "damage_multiplier", Type: "float32"},
			{GoName: "Delay", NBTName: "delay", Type: "int32"},
			{GoName: "DismountConditions", NBTName: "dismount_conditions", Type: "WeaponConditions", Codec: CodecNested, SliceElement: "WeaponConditions",
				Fields: []Field{
					{GoName: "MaxDuration", NBTName: "max_duration", Type: "int32"},
					{GoName: "MinSpeed", NBTName: "min_speed", Type: "float32"},
					{GoName: "MinRelativeSpeed", NBTName: "min_relative_speed", Type: "float32"},
				}},
			{GoName: "HitboxMargin", NBTName: "hitbox_margin", Type: "float32"},
			{GoName: "KnockbackConditions", NBTName: "knockback_conditions", Type: "WeaponConditions", Codec: CodecNested, SliceElement: "WeaponConditions",
				Fields: []Field{
					{GoName: "MaxDuration", NBTName: "max_duration", Type: "int32"},
					{GoName: "MinSpeed", NBTName: "min_speed", Type: "float32"},
					{GoName: "MinRelativeSpeed", NBTName: "min_relative_speed", Type: "float32"},
				}},
			{GoName: "Reach", NBTName: "reach", Type: "[2]float32", Codec: CodecRange},
		},
	},
	{
		Struct:  "UseModifiers",
		Name:    "minecraft:use_modifiers",
		Comment: "UseModifiers represents the minecraft:use_modifiers component.",
		Fields: []Field{
			{GoName: "MovementModifier", NBTName: "movement_modifier", Type: "float32"},
			{GoName: "UseDuration", NBTName: "use_duration", Type: "float32"},
			{GoName: "EmitVibrations", NBTName: "emit_vibrations", Type: "bool"},
			{GoName: "StartSound", NBTName: "start_sound", Type: "string", OmitEmpty: true},
			{GoName: "StartUsing", NBTName: "start_using", Type: "string", OmitEmpty: true},
		},
	},
	{
		Struct:  "StorageItem",
		Name:    "minecraft:storage_item",
		Comment: "StorageItem represents the minecraft:storage_item component.",
		Fields: []Field{
			{GoName: "AllowNestedStorageItems", NBTName: "allow_nested_storage_items", Type: "bool"},
			{GoName: "AllowedItems", NBTName: "allowed_items", Type: "[]string", Codec: CodecStringSlice},
			{GoName: "BannedItems", NBTName: "banned_items", Type: "[]string", Codec: CodecBanned},
			{GoName: "MaxSlots", NBTName: "max_slots", Type: "int32"},
		},
	},
	{
		Struct:  "Shooter",
		Name:    "minecraft:shooter",
		Comment: "Shooter represents the minecraft:shooter component.",
		Fields: []Field{
			{GoName: "Ammunition", NBTName: "ammunition", Type: "[]Ammunition", Codec: CodecSlice, SliceElement: "Ammunition",
				Fields: []Field{
					{GoName: "Item", NBTName: "item", Type: "string"},
					{GoName: "SearchInventory", NBTName: "search_inventory", Type: "bool"},
					{GoName: "UseInCreative", NBTName: "use_in_creative", Type: "bool"},
					{GoName: "UseOffHand", NBTName: "use_offhand", Type: "bool"},
				}},
			{GoName: "ChargeOnDraw", NBTName: "charge_on_draw", Type: "bool"},
			{GoName: "MaxDrawDuration", NBTName: "max_draw_duration", Type: "float32"},
			{GoName: "ScalePowerByDrawDuration", NBTName: "scale_power_by_draw_duration", Type: "bool"},
		},
	},
	{
		Struct:  "Projectile",
		Name:    "minecraft:projectile",
		Comment: "Projectile represents the minecraft:projectile component.",
		Fields: []Field{
			{GoName: "MinimumCriticalPower", NBTName: "minimum_critical_power", Type: "float32"},
			{GoName: "ProjectileEntity", NBTName: "projectile_entity", Type: "string"},
		},
	},
	{
		Struct:  "Throwable",
		Name:    "minecraft:throwable",
		Comment: "Throwable represents the minecraft:throwable component.",
		Fields: []Field{
			{GoName: "DoSwingAnimation", NBTName: "do_swing_animation", Type: "bool"},
			{GoName: "LaunchPowerScale", NBTName: "launch_power_scale", Type: "float32", OmitEmpty: true},
			{GoName: "MaxDrawDuration", NBTName: "max_draw_duration", Type: "float32", OmitEmpty: true},
			{GoName: "MaxLaunchPower", NBTName: "max_launch_power", Type: "float32", OmitEmpty: true},
			{GoName: "MinDrawDuration", NBTName: "min_draw_duration", Type: "float32", OmitEmpty: true},
			{GoName: "ScalePowerByDrawDuration", NBTName: "scale_power_by_draw_duration", Type: "bool", OmitEmpty: true},
		},
	},
	{
		Struct:  "Damage",
		Name:    "minecraft:damage",
		Comment: "Damage represents the minecraft:damage component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "float32"},
		},
	},
	{
		Struct:  "Cooldown",
		Name:    "minecraft:cooldown",
		Comment: "Cooldown represents the minecraft:cooldown component.",
		Fields: []Field{
			{GoName: "Category", NBTName: "category", Type: "string"},
			{GoName: "Duration", NBTName: "duration", Type: "float32"},
			{GoName: "Type", NBTName: "type", Type: "string", OmitEmpty: true},
		},
	},
	{
		Struct:  "Enchantable",
		Name:    "minecraft:enchantable",
		Comment: "Enchantable represents the minecraft:enchantable component.",
		Fields: []Field{
			{GoName: "Slot", NBTName: "slot", Type: "string"},
			{GoName: "Value", NBTName: "value", Type: "int32"},
		},
	},
	{
		Struct:  "Repairable",
		Name:    "minecraft:repairable",
		Comment: "Repairable represents the minecraft:repairable component.",
		Fields: []Field{
			{GoName: "RepairItems", NBTName: "repair_items", Type: "[]RepairEntry", Codec: CodecSlice, SliceElement: "RepairEntry",
				Fields: []Field{
					{GoName: "Items", NBTName: "items", Type: "[]RepairItemEntry", Codec: CodecNameOrTags, SliceElement: "RepairItemEntry",
						Fields: []Field{
							{GoName: "Name", NBTName: "name", Type: "string"},
							{GoName: "Tags", NBTName: "tags", Type: "[]string", Codec: CodecStringSlice},
						}},
					{GoName: "RepairAmount", NBTName: "repair_amount", Type: "int32"},
				}},
		},
	},
	{
		Struct:  "ItemTags",
		Name:    "minecraft:item_tags",
		Comment: "ItemTags represents the minecraft:item_tags component.",
		Fields: []Field{
			{GoName: "Tags", NBTName: "tags", Type: "[]string", Codec: CodecStringSlice},
		},
	},
	{
		Struct:  "Tags",
		Name:    "minecraft:tags",
		Comment: "Tags represents the minecraft:tags component.",
		Fields: []Field{
			{GoName: "Tags", NBTName: "tags", Type: "[]string", Codec: CodecStringSlice},
		},
	},
	{
		Struct:  "Seed",
		Name:    "minecraft:seed",
		Comment: "Seed represents the minecraft:seed component.",
		Fields: []Field{
			{GoName: "CropResult", NBTName: "crop_result", Type: "string"},
			{GoName: "PlantAt", NBTName: "plant_at", Type: "[]string", Codec: CodecStringSlice},
			{GoName: "PlantAtAnySolidSurface", NBTName: "plant_at_any_solid_surface", Type: "bool"},
			{GoName: "PlantAtFace", NBTName: "plant_at_face", Type: "string"},
		},
	},
	{
		Struct:  "Fuel",
		Name:    "minecraft:fuel",
		Comment: "Fuel represents the minecraft:fuel component.",
		Fields: []Field{
			{GoName: "Duration", NBTName: "duration", Type: "float32"},
		},
	},
	{
		Struct:  "FireResistant",
		Name:    "minecraft:fire_resistant",
		Comment: "FireResistant represents the minecraft:fire_resistant component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "bool"},
		},
	},
	{
		Struct:  "Glint",
		Name:    "minecraft:glint",
		Comment: "Glint represents the minecraft:glint component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "bool"},
		},
	},
	{
		Struct:  "SwingDuration",
		Name:    "minecraft:swing_duration",
		Comment: "SwingDuration represents the minecraft:swing_duration component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "float32"},
		},
	},
	{
		Struct:  "SwingSounds",
		Name:    "minecraft:swing_sounds",
		Comment: "SwingSounds represents the minecraft:swing_sounds component.",
		Fields: []Field{
			{GoName: "AttackCriticalHit", NBTName: "attack_critical_hit", Type: "string"},
			{GoName: "AttackHit", NBTName: "attack_hit", Type: "string"},
			{GoName: "AttackMiss", NBTName: "attack_miss", Type: "string"},
		},
	},
	{
		Struct:  "PiercingWeapon",
		Name:    "minecraft:piercing_weapon",
		Comment: "PiercingWeapon represents the minecraft:piercing_weapon component.",
		Fields: []Field{
			{GoName: "CreativeReach", NBTName: "creative_reach", Type: "[2]float32", Codec: CodecRange},
			{GoName: "HitboxMargin", NBTName: "hitbox_margin", Type: "float32"},
			{GoName: "Reach", NBTName: "reach", Type: "[2]float32", Codec: CodecRange},
		},
	},
	{
		Struct:  "Camera",
		Name:    "minecraft:camera",
		Comment: "Camera represents the minecraft:camera component.",
		Fields: []Field{
			{GoName: "BlackBarsDuration", NBTName: "black_bars_duration", Type: "float32"},
			{GoName: "BlackBarsScreenRatio", NBTName: "black_bars_screen_ratio", Type: "float32"},
			{GoName: "PictureDuration", NBTName: "picture_duration", Type: "float32"},
			{GoName: "ShutterDuration", NBTName: "shutter_duration", Type: "float32"},
			{GoName: "ShutterScreenRatio", NBTName: "shutter_screen_ratio", Type: "float32"},
			{GoName: "SlideAwayDuration", NBTName: "slide_away_duration", Type: "float32"},
			{GoName: "UseDuration", NBTName: "use_duration", Type: "int32"},
		},
	},
	{
		Struct:  "Block",
		Name:    "minecraft:block",
		Comment: "Block represents the minecraft:block component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "string"},
		},
	},
	{
		Struct:  "BlockPlacer",
		Name:    "minecraft:block_placer",
		Comment: "BlockPlacer represents the minecraft:block_placer component.",
		Fields: []Field{
			{GoName: "Block", NBTName: "block", Type: "string"},
			{GoName: "ReplaceBlockItem", NBTName: "replace_block_item", Type: "string"},
			{GoName: "AlignedPlacement", NBTName: "aligned_placement", Type: "bool"},
			{GoName: "UseOn", NBTName: "use_on", Type: "[]string", Codec: CodecStringSlice, OmitEmpty: true},
		},
	},
	{
		Struct:  "Compostable",
		Name:    "minecraft:compostable",
		Comment: "Compostable represents the minecraft:compostable component.",
		Fields: []Field{
			{GoName: "CompostingChance", NBTName: "composting_chance", Type: "int32"},
		},
	},
	{
		Struct:  "DamageAbsorption",
		Name:    "minecraft:damage_absorption",
		Comment: "DamageAbsorption represents the minecraft:damage_absorption component.",
		Fields: []Field{
			{GoName: "AbsorbableCauses", NBTName: "absorbable_causes", Type: "[]string", Codec: CodecStringSlice},
		},
	},
	{
		Struct:  "Digger",
		Name:    "minecraft:digger",
		Comment: "Digger represents the minecraft:digger component.",
		Fields: []Field{
			{GoName: "DestroySpeeds", NBTName: "destroy_speeds", Type: "[]DestroySpeed", Codec: CodecSlice, SliceElement: "DestroySpeed",
				Fields: []Field{
					{GoName: "Block", NBTName: "block", Type: "string"},
					{GoName: "Speed", NBTName: "speed", Type: "float32"},
				}},
			{GoName: "UseEfficiency", NBTName: "use_efficiency", Type: "bool"},
		},
	},
	{
		Struct:  "DurabilitySensor",
		Name:    "minecraft:durability_sensor",
		Comment: "DurabilitySensor represents the minecraft:durability_sensor component.",
		Fields: []Field{
			{GoName: "SoundEvent", NBTName: "sound_event", Type: "string", OmitEmpty: true},
			{GoName: "DurabilityThresholds", NBTName: "durability_thresholds", Type: "[]DurabilityThreshold", Codec: CodecSlice, SliceElement: "DurabilityThreshold", OmitEmpty: true,
				Fields: []Field{
					{GoName: "Durability", NBTName: "durability", Type: "int32"},
					{GoName: "ParticleType", NBTName: "particle_type", Type: "string", OmitEmpty: true},
					{GoName: "SoundEvent", NBTName: "sound_event", Type: "string", OmitEmpty: true},
				}},
		},
	},
	{
		Struct:  "Dyeable",
		Name:    "minecraft:dyeable",
		Comment: "Dyeable represents the minecraft:dyeable component.",
		Fields: []Field{
			{GoName: "DefaultColor", NBTName: "default_color", Type: "[3]int32", Codec: CodecList3},
		},
	},
	{
		Struct:  "EntityPlacer",
		Name:    "minecraft:entity_placer",
		Comment: "EntityPlacer represents the minecraft:entity_placer component.",
		Fields: []Field{
			{GoName: "Entity", NBTName: "entity", Type: "string"},
			{GoName: "UseOn", NBTName: "use_on", Type: "[]string", Codec: CodecStringSlice, OmitEmpty: true},
			{GoName: "DispenseOn", NBTName: "dispense_on", Type: "[]string", Codec: CodecStringSlice, OmitEmpty: true},
		},
	},
	{
		Struct:  "HoverTextColor",
		Name:    "minecraft:hover_text_color",
		Comment: "HoverTextColor represents the minecraft:hover_text_color component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "int32"},
		},
	},
	{
		Struct:  "InteractButton",
		Name:    "minecraft:interact_button",
		Comment: "InteractButton represents the minecraft:interact_button component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "string"},
		},
	},
	{
		Struct:  "LiquidClipped",
		Name:    "minecraft:liquid_clipped",
		Comment: "LiquidClipped represents the minecraft:liquid_clipped component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "bool"},
		},
	},
	{
		Struct:  "Rarity",
		Name:    "minecraft:rarity",
		Comment: "Rarity represents the minecraft:rarity component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "string"},
		},
	},
	{
		Struct:  "Record",
		Name:    "minecraft:record",
		Comment: "Record represents the minecraft:record component.",
		Fields: []Field{
			{GoName: "ComparatorSignal", NBTName: "comparator_signal", Type: "int32"},
			{GoName: "Duration", NBTName: "duration", Type: "float32"},
			{GoName: "SoundEvent", NBTName: "sound_event", Type: "string"},
		},
	},
	{
		Struct:  "ShouldDespawn",
		Name:    "minecraft:should_despawn",
		Comment: "ShouldDespawn represents the minecraft:should_despawn component.",
		Fields: []Field{
			{GoName: "Value", NBTName: "value", Type: "bool"},
		},
	},
	{
		Struct:  "BundleInteraction",
		Name:    "minecraft:bundle_interaction",
		Comment: "BundleInteraction represents the minecraft:bundle_interaction component.",
		Fields: []Field{
			{GoName: "NumViewableSlots", NBTName: "num_viewable_slots", Type: "int32"},
		},
	},
	{
		Struct:  "StorageWeightLimit",
		Name:    "minecraft:storage_weight_limit",
		Comment: "StorageWeightLimit represents the minecraft:storage_weight_limit component.",
		Fields: []Field{
			{GoName: "MaxWeightLimit", NBTName: "max_weight_limit", Type: "int32"},
		},
	},
	{
		Struct:  "StorageWeightModifier",
		Name:    "minecraft:storage_weight_modifier",
		Comment: "StorageWeightModifier represents the minecraft:storage_weight_modifier component.",
		Fields: []Field{
			{GoName: "WeightInStorageItem", NBTName: "weight_in_storage_item", Type: "int32"},
		},
	},
}
