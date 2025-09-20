package block

// EndBricks is a block made from combining four endstone blocks together.
type EndBricks struct {
	solid
	bassDrum
}

func (e EndBricks) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(e)).withBlastResistance(45)
}

func (EndBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:end_bricks", 0
}

func (EndBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_bricks", nil
}
