package block

// Bricks are decorative building blocks.
type Bricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (b Bricks) BreakInfo() BreakInfo {
	return NewBreakInfo(2, PickaxeHarvestable, PickaxeEffective, OneOf(b)).withBlastResistance(30)
}

// EncodeItem ...
func (Bricks) EncodeItem() (name string, meta int16) {
	return "minecraft:brick_block", 0
}

// EncodeBlock ...
func (Bricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:brick_block", nil
}
