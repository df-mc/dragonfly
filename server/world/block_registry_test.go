package world

import "testing"

// TestBlockInfoLookupsGuardOutOfRangeRuntimeID verifies that per-runtime-ID lookups return unknown-block defaults
// instead of panicking when handed a runtime ID beyond the registry, as happens when chunk decoding preserves an
// unknown network hash as a raw palette value.
func TestBlockInfoLookupsGuardOutOfRangeRuntimeID(t *testing.T) {
	DefaultBlockRegistry.Finalize()
	registry := DefaultBlockRegistry.Clone()

	rid := ^uint32(0)
	if got := registry.FilteringBlock(rid); got != 15 {
		t.Fatalf("FilteringBlock(%d) = %d, want 15", rid, got)
	}
	if got := registry.LightBlock(rid); got != 0 {
		t.Fatalf("LightBlock(%d) = %d, want 0", rid, got)
	}
	if registry.RandomTickBlock(rid) {
		t.Fatalf("RandomTickBlock(%d) = true, want false", rid)
	}
	if registry.NBTBlock(rid) {
		t.Fatalf("NBTBlock(%d) = true, want false", rid)
	}
	if registry.LiquidBlock(rid) {
		t.Fatalf("LiquidBlock(%d) = true, want false", rid)
	}
	if registry.LiquidDisplacingBlock(rid) {
		t.Fatalf("LiquidDisplacingBlock(%d) = true, want false", rid)
	}
}
