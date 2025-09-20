package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// GlassPane is a transparent block that can be used as a more efficient alternative to glass blocks.
type GlassPane struct {
	transparent
	thin
	clicksAndSticks
	sourceWaterDisplacer
}

func (p GlassPane) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (p GlassPane) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, silkTouchOnlyDrop(p))
}

func (GlassPane) EncodeItem() (name string, meta int16) {
	return "minecraft:glass_pane", meta
}

func (GlassPane) EncodeBlock() (string, map[string]any) {
	return "minecraft:glass_pane", nil
}
