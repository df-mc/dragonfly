package server

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestVanillaItemEntriesUseRegisteredMaxCount(t *testing.T) {
	srv := &Server{}

	tests := map[string]int32{
		"minecraft:ender_eye":   16,
		"minecraft:ender_pearl": 16,
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			entry, ok := vanillaItemEntry(name, srv.itemEntries())
			if !ok {
				t.Fatalf("expected item entry for %s", name)
			}
			got, ok := itemEntryMaxStackSize(entry.Data)
			if !ok {
				t.Fatalf("expected max_stack_size in item entry for %s", name)
			}
			if got != want {
				t.Fatalf("expected %s max_stack_size %d, got %d", name, want, got)
			}
		})
	}
}

func vanillaItemEntry(name string, entries []protocol.ItemEntry) (protocol.ItemEntry, bool) {
	for _, entry := range entries {
		if entry.Name == name {
			return entry, true
		}
	}
	return protocol.ItemEntry{}, false
}

func itemEntryMaxStackSize(data map[string]any) (int32, bool) {
	components, ok := data["components"].(map[string]any)
	if !ok {
		return 0, false
	}
	properties, ok := components["item_properties"].(map[string]any)
	if !ok {
		return 0, false
	}
	switch v := properties["max_stack_size"].(type) {
	case int32:
		return v, true
	case int:
		return int32(v), true
	case int16:
		return int32(v), true
	case int64:
		return int32(v), true
	default:
		return 0, false
	}
}
