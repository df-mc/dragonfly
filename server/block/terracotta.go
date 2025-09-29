package block

import "github.com/df-mc/dragonfly/server/world"

// Terracotta is a block formed from clay, with a hardness and blast resistance comparable to stone. For colouring it,
// take a look at StainedTerracotta.
type Terracotta struct {
	solid
	bassDrum
}

func (Terracotta) SoilFor(block world.Block) bool {
	_, ok := block.(DeadBush)
	return ok
}

func (t Terracotta) BreakInfo() BreakInfo {
	return newBreakInfo(1.25, pickaxeHarvestable, pickaxeEffective, oneOf(t)).withBlastResistance(21)
}

func (Terracotta) EncodeItem() (name string, meta int16) {
	return "minecraft:hardened_clay", meta
}

func (Terracotta) EncodeBlock() (string, map[string]any) {
	return "minecraft:hardened_clay", nil
}
