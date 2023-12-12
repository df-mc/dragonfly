package block

// PackedMud is a block crafted from mud and wheat. It is used to create mud bricks.
type PackedMud struct {
	solid
}

// BreakInfo ...
func (p PackedMud) BreakInfo() BreakInfo {
	return NewBreakInfo(1, AlwaysHarvestable, NothingEffective, OneOf(p)).withBlastResistance(15)
}

// EncodeItem ...
func (PackedMud) EncodeItem() (name string, meta int16) {
	return "minecraft:packed_mud", 0
}

// EncodeBlock ...
func (PackedMud) EncodeBlock() (string, map[string]any) {
	return "minecraft:packed_mud", nil
}
