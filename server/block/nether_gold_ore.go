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
	// TODO: Silk touch.
	i := newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, simpleDrops(item.NewStack(item.GoldNugget{}, rand.Intn(4)+2)))
	i.XPDrops = XPDropRange{0, 1}
	return i
}

// EncodeItem ...
func (NetherGoldOre) EncodeItem() (name string, meta int16) {
	return "minecraft:nether_gold_ore", 0
}

// EncodeBlock ...
func (NetherGoldOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:nether_gold_ore", nil
}
