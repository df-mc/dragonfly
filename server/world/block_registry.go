package world

import (
	"fmt"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/zaataylor/cartesian/cartesian"
)

func splitNamespace(identifier string) (ns, name string) {
	ns_name := strings.Split(identifier, ":")
	return ns_name[0], ns_name[len(ns_name)-1]
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
				enabled_states := trait["enabled_states"].(map[string]any)
				for k, enabled := range enabled_states {
					if !strings.ContainsRune(k, ':') {
						k = "minecraft:" + k
					}
					enabled := enabled.(uint8)
					if enabled == 0 {
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
			registerBlockState(blockState{
				Name:       entry.Name,
				Properties: m,
			})
		}
	}

	return nil
}
