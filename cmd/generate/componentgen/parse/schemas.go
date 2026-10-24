package parse

// ComponentSchema represents the schema of a component extracted from vanilla items.
type ComponentSchema struct {
	Name       string
	SampleData map[string]any
	ItemCount  int
}

// ExtractComponentSchemas extracts all unique component schemas from vanilla items.
func ExtractComponentSchemas(items map[string]VanillaItemEntry) map[string]ComponentSchema {
	schemas := make(map[string]ComponentSchema)

	for _, item := range items {
		if item.Data == nil {
			continue
		}
		components, ok := item.Data["components"].(map[string]any)
		if !ok {
			continue
		}
		for compName, compData := range components {
			if existing, ok := schemas[compName]; ok {
				existing.ItemCount++
				schemas[compName] = existing
			} else {
				dataMap, _ := compData.(map[string]any)
				schemas[compName] = ComponentSchema{
					Name:       compName,
					SampleData: dataMap,
					ItemCount:  1,
				}
			}
		}
	}

	return schemas
}

// AllComponentNames returns a sorted list of all component names.
func AllComponentNames(schemas map[string]ComponentSchema) []string {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	// Simple sort
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] > names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	return names
}

// ComponentSchemasToMap converts schemas to a map for template use.
func ComponentSchemasToMap(schemas map[string]ComponentSchema) map[string]any {
	result := make(map[string]any)
	for name, schema := range schemas {
		result[name] = map[string]any{
			"name":        schema.Name,
			"sample_data": schema.SampleData,
			"item_count":  schema.ItemCount,
		}
	}
	return result
}

// FilterKnownComponents filters to only known components that we have typed implementations for.
func FilterKnownComponents(schemas map[string]ComponentSchema) map[string]ComponentSchema {
	known := map[string]bool{
		"minecraft:wearable":                true,
		"minecraft:food":                    true,
		"minecraft:durability":              true,
		"minecraft:kinetic_weapon":          true,
		"minecraft:use_modifiers":           true,
		"minecraft:storage_item":            true,
		"minecraft:shooter":                 true,
		"minecraft:projectile":              true,
		"minecraft:throwable":               true,
		"minecraft:damage":                  true,
		"minecraft:cooldown":                true,
		"minecraft:enchantable":             true,
		"minecraft:repairable":              true,
		"minecraft:item_tags":               true,
		"minecraft:tags":                    true,
		"minecraft:seed":                    true,
		"minecraft:fuel":                    true,
		"minecraft:fire_resistant":          true,
		"minecraft:glint":                   true,
		"minecraft:swing_duration":          true,
		"minecraft:swing_sounds":            true,
		"minecraft:piercing_weapon":         true,
		"minecraft:camera":                  true,
		"minecraft:block":                   true,
		"minecraft:block_placer":            true,
		"minecraft:compostable":             true,
		"minecraft:damage_absorption":       true,
		"minecraft:digger":                  true,
		"minecraft:durability_sensor":       true,
		"minecraft:dyeable":                 true,
		"minecraft:entity_placer":           true,
		"minecraft:hover_text_color":        true,
		"minecraft:interact_button":         true,
		"minecraft:liquid_clipped":          true,
		"minecraft:rarity":                  true,
		"minecraft:record":                  true,
		"minecraft:should_despawn":          true,
		"minecraft:bundle_interaction":      true,
		"minecraft:storage_weight_limit":    true,
		"minecraft:storage_weight_modifier": true,
	}

	result := make(map[string]ComponentSchema)
	for name, schema := range schemas {
		if known[name] {
			result[name] = schema
		}
	}
	return result
}
