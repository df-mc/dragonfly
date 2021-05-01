package block

import (
	"github.com/df-mc/dragonfly/server/block/colour"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Concrete is a solid block which comes in the 16 regular dye colors, created by placing concrete powder
// adjacent to water.
type Concrete struct {
	solid
	bassDrum

	// Colour is the colour of the concrete block.
	Colour colour.Colour
}

// BreakInfo ...
func (c Concrete) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    1.8,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(c, 1)),
	}
}

// EncodeItem ...
func (c Concrete) EncodeItem() (id int32, name string, meta int16) {
	return 236, "minecraft:concrete", int16(c.Colour.Uint8())
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
