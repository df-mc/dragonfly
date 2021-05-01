package block

import (
	"github.com/df-mc/dragonfly/server/block/colour"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
)

// StainedGlass is a decorative, fully transparent solid block that is dyed into a different colour.
type StainedGlass struct {
	transparent
	solid
	clicksAndSticks

	// Colour specifies the colour of the block.
	Colour colour.Colour
}

// BreakInfo ...
func (g StainedGlass) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0.3,
		Harvestable: func(t tool.Tool) bool {
			return true // TODO(lhochbaum): Glass can be silk touched, implement silk touch.
		},
		Effective: nothingEffective,
		Drops:     simpleDrops(),
	}
}

// EncodeItem ...
func (g StainedGlass) EncodeItem() (id int32, name string, meta int16) {
	return 241, "minecraft:stained_glass", int16(g.Colour.Uint8())
}

// EncodeBlock ...
func (g StainedGlass) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:stained_glass", map[string]interface{}{"color": g.Colour.String()}
}

// allStainedGlass returns stained glass blocks with all possible colours.
func allStainedGlass() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range colour.All() {
		b = append(b, StainedGlass{Colour: c})
	}
	return b
}
