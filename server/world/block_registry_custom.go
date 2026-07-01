package world

import (
	"fmt"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/zaataylor/cartesian/cartesian"
)

func splitNamespace(identifier string) (ns, name string) {
	parts := strings.Split(identifier, ":")
	return parts[0], parts[len(parts)-1]
}

var traitLookup = map[string][]any{
	"minecraft:facing_direction": {
		"north", "east", "south", "west", "down", "up",
	},
	"minecraft:cardinal_direction": {
		"north", "east", "south", "west",
	},
	"minecraft:vertical_half": {
		"top", "bottom",
	},
	"minecraft:block_face": {
		"north", "east", "south", "west", "down", "up",
	},
}

// AddCustomBlocks registers each non-vanilla block entry's permutations as block states on the
// DefaultBlockRegistry so they can be resolved during chunk encoding/decoding.
func AddCustomBlocks(entries []protocol.BlockEntry) error {
	for _, entry := range entries {
		ns, _ := splitNamespace(entry.Name)
		if ns == "minecraft" {
			continue
		}

		var propertyNames []string
		var propertyValues []any

		props, ok := entry.Properties["properties"].([]any)
		if ok {
			for _, v := range props {
				v := v.(map[string]any)
				name := v["name"].(string)
				enum := v["enum"]
				propertyNames = append(propertyNames, name)
				propertyValues = append(propertyValues, enum)
			}
		}

		traits, ok := entry.Properties["traits"].([]any)
		if ok {
			for _, trait := range traits {
				trait := trait.(map[string]any)
				enabledStates := trait["enabled_states"].(map[string]any)
				for k, enabled := range enabledStates {
					if !strings.ContainsRune(k, ':') {
						k = "minecraft:" + k
					}
					if enabled.(uint8) == 0 {
						continue
					}
					v, ok := traitLookup[k]
					if !ok {
						return fmt.Errorf("unresolved trait %s", k)
					}

					propertyNames = append(propertyNames, k)
					propertyValues = append(propertyValues, v)
				}
			}
		}

		permutations := cartesian.NewCartesianProduct(propertyValues).Values()

		for _, values := range permutations {
			m := make(map[string]any)
			for i, value := range values {
				name := propertyNames[i]
				m[name] = value
			}
			// TODO: this is wrong, check codex/custom-block-registries branch
			DefaultBlockRegistry.RegisterBlockState(BlockState{
				Name:       entry.Name,
				Properties: m,
			})
		}
	}

	return nil
}
