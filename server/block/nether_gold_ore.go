package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"math/rand"
)

// NetherGoldOre is a variant of gold ore found exclusively in The Nether.
type NetherGoldOre struct {
	solid
}

// BreakInfo ...
func (n NetherGoldOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(item.GoldNugget{}, rand.Intn(4)+2)), //TODO: Silk Touch
		XPDrops:     XPDropRange{0, 1},
	}
}

// EncodeItem ...
func (NetherGoldOre) EncodeItem() (id int32, name string, meta int16) {
	return -288, "minecraft:nether_gold_ore", 0
}

// EncodeBlock ...
func (NetherGoldOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:nether_gold_ore", nil
}
