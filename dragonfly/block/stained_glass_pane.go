package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// StainedGlassPane is a transparent block that can be used as a more efficient alternative to glass blocks.
type StainedGlassPane struct {
	noNBT
	transparent
	thin

	// Colour specifies the colour of the block.
	Colour colour.Colour
}

// CanDisplace ...
func (p StainedGlassPane) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (p StainedGlassPane) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
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
func (p StainedGlassPane) EncodeItem() (id int32, meta int16) {
	return 160, int16(p.Colour.Uint8())
}

// EncodeBlock ...
func (p StainedGlassPane) EncodeBlock() (name string, properties map[string]interface{}) {
	colourName := p.Colour.String()
	if p.Colour == colour.LightGrey() {
		// And here we go again. Light grey is actually called "silver".
		colourName = "silver"
	}
	return "minecraft:stained_glass_pane", map[string]interface{}{"color": colourName}
}

// Hash ...
func (p StainedGlassPane) Hash() uint64 {
	return hashStainedGlassPane | uint64(p.Colour.Uint8())<<34
}

// allStainedGlassPane returns stained glass panes with all possible colours.
func allStainedGlassPane() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range colour.All() {
		b = append(b, StainedGlassPane{Colour: c})
	}
	return b
}
