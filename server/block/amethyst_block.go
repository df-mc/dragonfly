package block

// AmethystBlock is a decorative block crafted from four amethyst shards.
type AmethystBlock struct {
	solid
}

// BreakInfo ...
func (a AmethystBlock) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeHarvestable, oneOf(a))
}

// EncodeItem ...
func (AmethystBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:amethyst_block", 0
}

// EncodeBlock ...
func (AmethystBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:amethyst_block", nil
}
