package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/instrument"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// EmeraldBlock is a precious mineral block crafted using 9 emeralds.
type EmeraldBlock struct {
	noNBT
	solid
}

// Instrument ...
func (e EmeraldBlock) Instrument() instrument.Instrument {
	return instrument.Bit()
}

// BreakInfo ...
func (e EmeraldBlock) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 5,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(e, 1)),
	}
}

// PowersBeacon ...
func (EmeraldBlock) PowersBeacon() bool {
	return true
}

// EncodeItem ...
func (EmeraldBlock) EncodeItem() (id int32, meta int16) {
	return 133, 0
}

// EncodeBlock ...
func (EmeraldBlock) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:emerald_block", nil
}

// Hash ...
func (EmeraldBlock) Hash() uint64 {
	return hashEmeraldBlock
}
