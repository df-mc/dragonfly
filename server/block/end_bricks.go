package block

// EndBricks is a block made from combining four endstone blocks together.
type EndBricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (e EndBricks) BreakInfo() BreakInfo {
	return newBreakInfo(0.8, pickaxeHarvestable, pickaxeEffective, oneOf(e))
}

// EncodeItem ...
func (EndBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:end_bricks", 0
}

// EncodeBlock ...
func (EndBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_bricks", nil
}
