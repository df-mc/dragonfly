package world

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

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
	"minecraft:corner_and_cardinal_direction": {
		"none", "inner_left", "inner_right", "outer_left", "outer_right",
	},
}

const maxCustomBlockStates = 1 << 16

type customBlockStateSpace struct {
	name       string
	properties []string
	values     [][]any
}

// NewCustomBlockRegistry returns an independent registry with default block runtime IDs preserved and custom block
// states appended after the default registry.
func NewCustomBlockRegistry(entries []protocol.BlockEntry) (BlockRegistry, error) {
	DefaultBlockRegistry.Finalize()
	registry := DefaultBlockRegistry.Clone()
	if err := AddCustomBlocks(registry, entries); err != nil {
		return nil, err
	}
	return registry, nil
}

// AddCustomBlocks appends custom block states to registry while preserving its existing runtime IDs.
func AddCustomBlocks(registry BlockRegistry, entries []protocol.BlockEntry) error {
	basicRegistry, ok := registry.(*BasicBlockRegistry)
	if !ok {
		return fmt.Errorf("unsupported block registry type %T", registry)
	}
	spaces := make([]customBlockStateSpace, 0, len(entries))
	total := 0
	for _, entry := range entries {
		if namespace, _, _ := strings.Cut(entry.Name, ":"); namespace == "minecraft" {
			continue
		}

		propertyNames, propertyValues, err := customBlockPropertySpace(entry)
		if err != nil {
			return fmt.Errorf("custom block %s: %w", entry.Name, err)
		}
		count := 1
		for _, values := range propertyValues {
			if count > (maxCustomBlockStates-total)/len(values) {
				return fmt.Errorf("custom block %s: states exceed limit of %d", entry.Name, maxCustomBlockStates)
			}
			count *= len(values)
		}
		if count > maxCustomBlockStates-total {
			return fmt.Errorf("custom block %s: states exceed limit of %d", entry.Name, maxCustomBlockStates)
		}
		total += count
		spaces = append(spaces, customBlockStateSpace{entry.Name, propertyNames, propertyValues})
	}

	scratch := make([]byte, 0, 0xff)
	for _, space := range spaces {
		err := forEachCustomBlockState(space.properties, space.values, func(properties map[string]any) error {
			state := BlockState{Name: space.name, Properties: properties}
			var stateErr error
			scratch, stateErr = basicRegistry.addCustomBlockState(state, scratch)
			return stateErr
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func customBlockPropertySpace(entry protocol.BlockEntry) ([]string, [][]any, error) {
	var propertyNames []string
	var propertyValues [][]any

	props, err := customBlockMaps(entry.Properties["properties"], "properties")
	if err != nil {
		return nil, nil, err
	}
	for _, v := range props {
		name, ok := v["name"].(string)
		if !ok {
			return nil, nil, fmt.Errorf("expected property name to be string, got %T", v["name"])
		}
		enum, ok := v["enum"].([]any)
		if !ok {
			return nil, nil, fmt.Errorf("expected property %s enum to be []any, got %T", name, v["enum"])
		}
		if len(enum) == 0 {
			return nil, nil, fmt.Errorf("expected property %s enum to contain at least one value", name)
		}
		propertyNames = append(propertyNames, name)
		propertyValues = append(propertyValues, enum)
	}

	traits, err := customBlockMaps(entry.Properties["traits"], "traits")
	if err != nil {
		return nil, nil, err
	}
	for _, trait := range traits {
		enabledStates, ok := trait["enabled_states"].(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("expected enabled_states to be map[string]any, got %T", trait["enabled_states"])
		}
		keys := make([]string, 0, len(enabledStates))
		for k := range enabledStates {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			rawEnabled := enabledStates[k]
			if !strings.ContainsRune(k, ':') {
				k = "minecraft:" + k
			}
			enabled, err := customBlockTraitEnabled(rawEnabled)
			if err != nil {
				return nil, nil, fmt.Errorf("trait %s: %w", k, err)
			}
			if !enabled {
				continue
			}
			v, ok := traitLookup[k]
			if !ok {
				return nil, nil, fmt.Errorf("unresolved trait %s", k)
			}

			propertyNames = append(propertyNames, k)
			propertyValues = append(propertyValues, v)
		}
	}

	return propertyNames, propertyValues, nil
}

func customBlockMaps(value any, field string) ([]map[string]any, error) {
	switch value := value.(type) {
	case nil:
		return nil, nil
	case []map[string]any:
		return value, nil
	case []any:
		values := make([]map[string]any, len(value))
		for i, entry := range value {
			mapped, ok := entry.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("expected %s entry at index %d to be map[string]any, got %T", field, i, entry)
			}
			values[i] = mapped
		}
		return values, nil
	default:
		return nil, fmt.Errorf("expected %s to be a list of maps, got %T", field, value)
	}
}

func forEachCustomBlockState(names []string, valueSets [][]any, yield func(map[string]any) error) error {
	properties := make(map[string]any, len(names))
	var visit func(int) error
	visit = func(index int) error {
		if index == len(names) {
			state := make(map[string]any, len(properties))
			for name, value := range properties {
				state[name] = value
			}
			return yield(state)
		}
		for _, value := range valueSets[index] {
			properties[names[index]] = value
			if err := visit(index + 1); err != nil {
				return err
			}
		}
		return nil
	}
	return visit(0)
}

func customBlockTraitEnabled(v any) (bool, error) {
	switch v := v.(type) {
	case bool:
		return v, nil
	case uint8:
		return v != 0, nil
	case int32:
		return v != 0, nil
	default:
		return false, fmt.Errorf("expected enabled flag to be bool, uint8, or int32, got %T", v)
	}
}
