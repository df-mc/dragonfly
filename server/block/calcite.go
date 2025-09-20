package block

// Calcite is a carbonate mineral found as part of amethyst geodes.
type Calcite struct {
	solid
	bassDrum
}

func (c Calcite) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, pickaxeHarvestable, pickaxeEffective, oneOf(c))
}

func (c Calcite) EncodeItem() (name string, meta int16) {
	return "minecraft:calcite", 0
}

func (c Calcite) EncodeBlock() (string, map[string]any) {
	return "minecraft:calcite", nil
}
