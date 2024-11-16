package block

// PolishedTuff is a decorational variant of Tuff that can be crafted or found naturally in Trial Chambers.
type PolishedTuff struct {
	solid
	bassDrum
}

// BreakInfo ...
func (t PolishedTuff) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(t)).withBlastResistance(30)
}

// EncodeItem ...
func (t PolishedTuff) EncodeItem() (name string, meta int16) {
	return "minecraft:polished_tuff", 0
}

// EncodeBlock ...
func (t PolishedTuff) EncodeBlock() (string, map[string]any) {
	return "minecraft:polished_tuff", nil
}
