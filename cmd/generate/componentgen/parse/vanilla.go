package parse

import (
	"os"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

type VanillaItemEntry struct {
	RuntimeID      int32          `nbt:"runtime_id"`
	ComponentBased bool           `nbt:"component_based"`
	Version        int32          `nbt:"version"`
	Data           map[string]any `nbt:"data,omitempty"`
}

// VanillaItems parses the vanilla_items.nbt file and returns a map of item name to entry.
func VanillaItems(path string) (map[string]VanillaItemEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items map[string]VanillaItemEntry
	if err := nbt.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}
