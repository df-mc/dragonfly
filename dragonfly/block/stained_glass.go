package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// StainedGlass is a decorative, fully transparent solid block that is dyed into a different colour.
type StainedGlass struct {
	noNBT
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
func (g StainedGlass) EncodeItem() (id int32, meta int16) {
	return 241, int16(g.Colour.Uint8())
}

// EncodeBlock ...
func (g StainedGlass) EncodeBlock() (name string, properties map[string]interface{}) {
	colourName := g.Colour.String()
	if g.Colour == colour.LightGrey() {
		// And here we go again. Light grey is actually called "silver".
		colourName = "silver"
	}
	return "minecraft:stained_glass", map[string]interface{}{"color": colourName}
}

// Hash ...
func (g StainedGlass) Hash() uint64 {
	return hashStainedGlass | uint64(g.Colour.Uint8())<<34
}
