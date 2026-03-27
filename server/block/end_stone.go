package block

import "github.com/df-mc/dragonfly/server/world"

// EndStone is a block found in The End.
type EndStone struct {
	solid
	bassDrum
}

// SoilFor ...
func (e EndStone) SoilFor(b world.Block) bool {
	switch b.(type) {
	case ChorusPlant, ChorusFlower:
		return true
	default:
		return false
	}
}

// BreakInfo ...
func (e EndStone) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(e)).withBlastResistance(45)
}

// EncodeItem ...
func (EndStone) EncodeItem() (name string, meta int16) {
	return "minecraft:end_stone", 0
}

// EncodeBlock ...
func (EndStone) EncodeBlock() (string, map[string]any) {
	return "minecraft:end_stone", nil
}
