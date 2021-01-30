package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// DiamondBlock is a block which can only be gained by crafting it.
type DiamondBlock struct {
	noNBT
	solid
}

// BreakInfo ...
func (d DiamondBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 5,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(d, 1)),
	}
}

// PowersBeacon ...
func (DiamondBlock) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (DiamondBlock) EncodeItem() (id int32, meta int16) {
	return 57, 0
}
