package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// StainedGlassPane is a transparent block that can be used as a more efficient alternative to glass blocks.
type StainedGlassPane struct {
	transparent
	thin
	clicksAndSticks

	// Colour specifies the colour of the block.
	Colour item.Colour
}

// CanDisplace ...
func (p StainedGlassPane) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (p StainedGlassPane) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// BreakInfo ...
func (p StainedGlassPane) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, silkTouchOnlyDrop(p))
}

// EncodeItem ...
func (p StainedGlassPane) EncodeItem() (name string, meta int16) {
	return "minecraft:stained_glass_pane", int16(p.Colour.Uint8())
}

// EncodeBlock ...
func (p StainedGlassPane) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:stained_glass_pane", map[string]any{"color": p.Colour.String()}
}

// allStainedGlassPane returns stained-glass panes with all possible colours.
func allStainedGlassPane() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range item.Colours() {
		b = append(b, StainedGlassPane{Colour: c})
	}
	return b
}
