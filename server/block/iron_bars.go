package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// IronBars are blocks that serve a similar purpose to glass panes, but made of iron instead of glass.
type IronBars struct {
	transparent
	thin
	sourceWaterDisplacer
}

func (i IronBars) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(i)).withBlastResistance(30)
}

func (i IronBars) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (IronBars) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_bars", 0
}

func (IronBars) EncodeBlock() (string, map[string]any) {
	return "minecraft:iron_bars", nil
}
