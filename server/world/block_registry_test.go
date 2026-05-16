package world

import "testing"

func TestBasicBlockRegistryRuntimeIDToHashRoundTrip(t *testing.T) {
	DefaultBlockRegistry.Finalize()

	blocks := DefaultBlockRegistry.Blocks()
	for runtimeID := range blocks {
		hash, ok := DefaultBlockRegistry.RuntimeIDToHash(uint32(runtimeID))
		if !ok {
			t.Fatalf("expected network hash for runtime ID %d", runtimeID)
		}
		gotRuntimeID, ok := DefaultBlockRegistry.HashToRuntimeID(hash)
		if !ok {
			t.Fatalf("expected runtime ID for network hash %d", hash)
		}
		if gotRuntimeID != uint32(runtimeID) {
			t.Fatalf("expected runtime ID %d for network hash %d, got %d", runtimeID, hash, gotRuntimeID)
		}
	}

	if _, ok := DefaultBlockRegistry.RuntimeIDToHash(uint32(len(blocks))); ok {
		t.Fatalf("expected no network hash for out-of-range runtime ID %d", len(blocks))
	}
}
