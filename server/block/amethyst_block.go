package block

// AmethystBlock is a decorative block crafted from four amethyst shards.
type AmethystBlock struct {
	solid
}

// BreakInfo ...
func (a AmethystBlock) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeHarvestable, oneOf(a))
}

// Resistance ...
func (a AmethystBlock) Resistance() float64 {
	return 1.5
}

// AlwaysExplodeDrop ...
func (a AmethystBlock) AlwaysExplodeDrop() bool {
	return false
}

// EncodeItem ...
func (AmethystBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:amethyst_block", 0
}

// EncodeBlock ...
func (AmethystBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:amethyst_block", nil
}
