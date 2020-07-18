package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// Terracotta is a block formed from clay, with a hardness and blast resistance comparable to stone. For colouring it,
// take a look at StainedTerracotta.
type Terracotta struct{}

// BreakInfo ...
func (t Terracotta) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1.25,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(t, 1)),
	}
}

// EncodeItem ...
func (t Terracotta) EncodeItem() (id int32, meta int16) {
	return 172, meta
}

// EncodeBlock ...
func (t Terracotta) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:hardened_clay", map[string]interface{}{}
}

// Hash ...
func (t Terracotta) Hash() uint64 {
	return hashTerracotta
}
