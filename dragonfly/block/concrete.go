package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/colour"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Concrete is a solid block which comes in the 16 regular dye colors, created by placing concrete powder
// adjacent to water.
type Concrete struct {
	// Colour is the colour of the concrete block.
	Colour colour.Colour
}

// EncodeItem ...
func (c Concrete) EncodeItem() (id int32, meta int16) {
	return 236, int16(c.Colour.Uint8())
}

// EncodeBlock ...
func (c Concrete) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:concrete", map[string]interface{}{"color": c.Colour.String()}
}

// allConcrete returns concrete blocks with all possible colours.
func allConcrete() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range colour.All() {
		b = append(b, Concrete{Colour: c})
	}
	return b
}
