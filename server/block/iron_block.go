package block

import (
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// IronBlock is a precious metal block made from 9 iron ingots.
type IronBlock struct {
	solid
}

// Instrument ...
func (i IronBlock) Instrument() instrument.Instrument {
	return instrument.IronXylophone()
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
func (IronBlock) EncodeItem() (id int32, name string, meta int16) {
	return 42, "minecraft:iron_block", 0
}

// EncodeBlock ...
func (IronBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:iron_block", nil
}
