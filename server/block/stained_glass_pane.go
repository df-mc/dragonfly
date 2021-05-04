package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
)

// StainedGlassPane is a transparent block that can be used as a more efficient alternative to glass blocks.
type StainedGlassPane struct {
	transparent
	thin
	clicksAndSticks

	// Colour specifies the colour of the block.
	Colour Colour
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
	return BreakInfo{
		Hardness: 0.3,
		Harvestable: func(t tool.Tool) bool {
			return true // TODO(lhochbaum): Glass panes can be silk touched, implement silk touch.
		},
		Effective: nothingEffective,
		Drops:     simpleDrops(),
	}
}

// EncodeItem ...
func (p StainedGlassPane) EncodeItem() (name string, meta int16) {
	return "minecraft:stained_glass_pane", int16(p.Colour.Uint8())
}

// EncodeBlock ...
func (p StainedGlassPane) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:stained_glass_pane", map[string]interface{}{"color": p.Colour.String()}
}

// allStainedGlassPane returns stained glass panes with all possible colours.
func allStainedGlassPane() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range Colours() {
		b = append(b, StainedGlassPane{Colour: c})
	}
	return b
}
