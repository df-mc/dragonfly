package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// IronBlock is a precious metal block made from 9 iron ingots.
type IronBlock struct {
	noNBT
	solid
}

// BreakInfo ...
func (i IronBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 5,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(i, 1)),
	}
}

// PowersBeacon ...
func (IronBlock) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (IronBlock) EncodeItem() (id int32, meta int16) {
	return 42, 0
}

// EncodeBlock ...
func (IronBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:iron_block", nil
}

// Hash ...
func (IronBlock) Hash() uint64 {
	return hashIronBlock
}
