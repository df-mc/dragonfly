package iteminternal

import (
	"fmt"
)

// ValidateComponents checks cross-component constraints derived from BDS 1.26.40.8
func ValidateComponents(components map[string]any) error {
	// kinetic_weapon requires use_modifiers
	if _, ok := components["minecraft:kinetic_weapon"]; ok {
		if _, ok := components["minecraft:use_modifiers"]; !ok {
			return fmt.Errorf("minecraft:kinetic_weapon requires minecraft:use_modifiers component")
		}
		// Validate kinetic_weapon conditions
		if err := validateKineticWeapon(components["minecraft:kinetic_weapon"]); err != nil {
			return err
		}
	}

	// bundle_interaction requires storage_item
	if _, ok := components["minecraft:bundle_interaction"]; ok {
		if _, ok := components["minecraft:storage_item"]; !ok {
			return fmt.Errorf("minecraft:bundle_interaction requires minecraft:storage_item component")
		}
	}

	// shooter requires non-zero use_duration (checked via use_modifiers component)
	if _, ok := components["minecraft:shooter"]; ok {
		if um, ok := components["minecraft:use_modifiers"].(map[string]any); ok {
			if ud, ok := um["use_duration"].(float32); !ok || ud == 0 {
				return fmt.Errorf("minecraft:shooter requires non-zero use_duration in minecraft:use_modifiers")
			}
		} else {
			return fmt.Errorf("minecraft:shooter requires minecraft:use_modifiers component with non-zero use_duration")
		}
	}

	// repairable validation: repair_items must be structured entries
	if rep, ok := components["minecraft:repairable"].(map[string]any); ok {
		if items, ok := rep["repair_items"].([]any); ok {
			for _, item := range items {
				if entry, ok := item.(map[string]any); ok {
					if _, hasItems := entry["items"]; !hasItems {
						return fmt.Errorf("minecraft:repairable repair_items must have 'items' field")
					}
				}
			}
		}
	}

	return nil
}

func validateKineticWeapon(kw any) error {
	data, ok := kw.(map[string]any)
	if !ok {
		return fmt.Errorf("kinetic_weapon must be a compound")
	}

	// At least one kinetic condition must be defined with max_duration > 0
	hasCondition := false
	for _, key := range []string{"damage_conditions", "knockback_conditions", "dismount_conditions"} {
		if cond, ok := data[key].(map[string]any); ok {
			if maxDur, ok := cond["max_duration"].(int32); ok && maxDur > 0 {
				hasCondition = true
				// min_speed and min_relative_speed are mutually exclusive
				_, hasMinSpeed := cond["min_speed"]
				_, hasMinRelSpeed := cond["min_relative_speed"]
				if hasMinSpeed && hasMinRelSpeed {
					return fmt.Errorf("kinetic_weapon %s: min_speed and min_relative_speed are mutually exclusive", key)
				}
			}
		}
	}
	if !hasCondition {
		return fmt.Errorf("kinetic_weapon: at least one condition with max_duration > 0 required")
	}

	return nil
}
