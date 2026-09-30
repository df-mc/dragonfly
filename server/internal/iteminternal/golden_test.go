package iteminternal

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
	"github.com/stretchr/testify/require"
)

func TestGoldenNBT(t *testing.T) {
	var vanillaItems map[string]VanillaItemEntry
	_ = nbt.Unmarshal(VanillaItemsData(), &vanillaItems)

	// Test a few key items that use the new component system
	testItems := []string{
		"minecraft:diamond_sword",
		"minecraft:diamond_spear",
		"minecraft:bow",
		"minecraft:crossbow",
		"minecraft:bundle",
		"minecraft:shield",
	}

	for _, name := range testItems {
		t.Run(name, func(t *testing.T) {
			item, ok := vanillaItems[name]
			require.True(t, ok, "item %s not found in vanilla items", name)

			if item.Data == nil {
				t.Skip("item has no component data")
			}

			components, ok := item.Data["components"].(map[string]any)
			if !ok {
				t.Skipf("item %s has no components map", name)
			}

			// Verify key components are present
			expectedComponents := expectedComponentsForItem(name)
			for _, compName := range expectedComponents {
				require.Contains(t, components, compName, "item %s missing component %s", name, compName)
			}

			// Log the component structure for debugging
			b, _ := json.MarshalIndent(components, "", "  ")
			t.Logf("Components for %s:\n%s", name, string(b))
		})
	}
}

type VanillaItemEntry struct {
	RuntimeID      int32          `nbt:"runtime_id"`
	ComponentBased bool           `nbt:"component_based"`
	Version        int32          `nbt:"version"`
	Data           map[string]any `nbt:"data,omitempty"`
}

func expectedComponentsForItem(name string) []string {
	switch name {
	case "minecraft:diamond_sword":
		return []string{"minecraft:damage", "minecraft:durability", "minecraft:enchantable", "minecraft:repairable"}
	case "minecraft:diamond_spear":
		return []string{"minecraft:kinetic_weapon", "minecraft:use_modifiers", "minecraft:damage", "minecraft:durability", "minecraft:enchantable", "minecraft:repairable"}
	case "minecraft:bow":
		return []string{"minecraft:projectile", "minecraft:durability", "minecraft:enchantable", "minecraft:repairable"}
	case "minecraft:crossbow":
		return []string{"minecraft:projectile", "minecraft:durability", "minecraft:enchantable", "minecraft:repairable", "minecraft:cooldown"}
	case "minecraft:bundle":
		return []string{"minecraft:bundle_interaction", "minecraft:storage_item"}
	case "minecraft:shield":
		return []string{"minecraft:durability", "minecraft:enchantable", "minecraft:repairable"}
	default:
		return nil
	}
}

func VanillaItemsData() []byte {
	data, _ := os.ReadFile("../../../server/world/vanilla_items.nbt")
	return data
}
