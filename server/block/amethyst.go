package block

// Amethyst is a decorative block crafted from four amethyst shards.
type Amethyst struct {
	solid
}

// BreakInfo ...
func (a Amethyst) BreakInfo() BreakInfo {
	return NewBreakInfo(1.5, PickaxeHarvestable, PickaxeHarvestable, OneOf(a))
}

// EncodeItem ...
func (Amethyst) EncodeItem() (name string, meta int16) {
	return "minecraft:amethyst_block", 0
}

// EncodeBlock ...
func (Amethyst) EncodeBlock() (string, map[string]any) {
	return "minecraft:amethyst_block", nil
}
