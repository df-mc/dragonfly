package block

import (
	"github.com/df-mc/dragonfly/server/item/tool"
)

// RawIronBlock is a raw metal block equivalent to nine raw iron.
type RawIronBlock struct {
	solid
	bassDrum
}

// BreakInfo ...
func (r RawIronBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, oneOf(r), XPDropRange{})
}

// EncodeItem ...
func (RawIronBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_iron_block", 0
}

// EncodeBlock ...
func (RawIronBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:raw_iron_block", nil
}
