package world

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func TestAddCustomBlocksRejectsExcessStates(t *testing.T) {
	properties := make([]any, 17)
	for i := range properties {
		properties[i] = map[string]any{
			"name": fmt.Sprintf("test:property_%d", i),
			"enum": []any{false, true},
		}
	}
	_, err := NewCustomBlockRegistry([]protocol.BlockEntry{{
		Name:       "test:block",
		Properties: map[string]any{"properties": properties},
	}})
	if err == nil || !strings.Contains(err.Error(), "exceed limit") {
		t.Fatalf("NewCustomBlockRegistry() error = %v, want state limit error", err)
	}
}

func TestAddCustomBlocksPreservesVanillaRuntimeIDs(t *testing.T) {
	DefaultBlockRegistry.Finalize()
	registry := DefaultBlockRegistry.Clone()

	oldCount := registry.BlockCount()
	oldAir := registry.AirRuntimeID()

	entry := protocol.BlockEntry{
		Name: "dragonfly:test_block",
		Properties: map[string]any{
			"properties": []any{
				map[string]any{
					"name": "dragonfly:variant",
					"enum": []any{int32(0), int32(1)},
				},
			},
		},
	}
	if err := AddCustomBlocks(registry, []protocol.BlockEntry{entry}); err != nil {
		t.Fatalf("AddCustomBlocks() error = %v", err)
	}

	if air := registry.AirRuntimeID(); air != oldAir {
		t.Fatalf("AirRuntimeID() changed: got %d, want %d", air, oldAir)
	}
	rid, ok := registry.StateToRuntimeID("dragonfly:test_block", map[string]any{"dragonfly:variant": int32(0)})
	if !ok {
		t.Fatalf("StateToRuntimeID() ok = false, want true")
	}
	if rid != uint32(oldCount) {
		t.Fatalf("StateToRuntimeID() rid = %d, want %d (custom states append after vanilla)", rid, oldCount)
	}
	if got, want := registry.BlockCount(), oldCount+2; got != want {
		t.Fatalf("BlockCount() = %d, want %d", got, want)
	}
}

func TestAddCustomBlocksSkipsDuplicateStates(t *testing.T) {
	DefaultBlockRegistry.Finalize()
	registry := DefaultBlockRegistry.Clone()

	entry := protocol.BlockEntry{Name: "dragonfly:plain_block", Properties: map[string]any{}}
	if err := AddCustomBlocks(registry, []protocol.BlockEntry{entry}); err != nil {
		t.Fatalf("first AddCustomBlocks() error = %v", err)
	}
	count := registry.BlockCount()
	if err := AddCustomBlocks(registry, []protocol.BlockEntry{entry}); err != nil {
		t.Fatalf("second AddCustomBlocks() error = %v", err)
	}
	if got := registry.BlockCount(); got != count {
		t.Fatalf("BlockCount() after duplicate = %d, want %d", got, count)
	}
}

func TestNewCustomBlockRegistryPreservesVanillaRuntimeIDs(t *testing.T) {
	DefaultBlockRegistry.Finalize()

	registry, err := NewCustomBlockRegistry([]protocol.BlockEntry{{Name: "dragonfly:plain_block"}})
	if err != nil {
		t.Fatalf("NewCustomBlockRegistry() error = %v", err)
	}
	basic, ok := registry.(*BasicBlockRegistry)
	if !ok {
		t.Fatalf("NewCustomBlockRegistry() returned %T, want *BasicBlockRegistry", registry)
	}
	if got, want := basic.AirRuntimeID(), DefaultBlockRegistry.AirRuntimeID(); got != want {
		t.Fatalf("AirRuntimeID() = %d, want %d", got, want)
	}
}
