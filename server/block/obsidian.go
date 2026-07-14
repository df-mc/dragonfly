package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Obsidian is a dark purple block known for its high blast resistance and strength, most commonly found when
// water flows over lava.
type Obsidian struct {
	solid
	bassDrum
	// Crying specifies if the block is a crying obsidian block. If true, the block is blue and emits light.
	Crying bool
}

// LightEmissionLevel ...
func (o Obsidian) LightEmissionLevel() uint8 {
	if o.Crying {
		return 10
	}
	return 0
}

// EncodeItem ...
func (o Obsidian) EncodeItem() (name string, meta int16) {
	if o.Crying {
		return "minecraft:crying_obsidian", 0
	}
	return "minecraft:obsidian", 0
}

// EncodeBlock ...
func (o Obsidian) EncodeBlock() (string, map[string]any) {
	if o.Crying {
		return "minecraft:crying_obsidian", nil
	}
	return "minecraft:obsidian", nil
}

// Frame returns true if the block can form part of a nether portal frame in the dimension passed. Crying obsidian
// cannot be used as a frame block.
func (o Obsidian) Frame(dimension world.Dimension) bool {
	return dimension == world.Nether && !o.Crying
}

// BreakInfo ...
func (o Obsidian) BreakInfo() BreakInfo {
	return newBreakInfo(35, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierDiamond.HarvestLevel
	}, pickaxeEffective, oneOf(o)).withBlastResistance(1200)
}
