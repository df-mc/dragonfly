package block

import "github.com/df-mc/dragonfly/dragonfly/item/tool"

// TODO: Slipperiness and SilkTouch

// Packedice is a solid block similar to ice
type PackedIce struct {
	noNBT
	solid
}

// BreakInfo ...
func (i PackedIce) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0.5,
		Harvestable: func(t tool.Tool) bool {
			return false //TODO: Silk touch
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(),
	}
}

// EncodeItem ...
func (PackedIce) EncodeItem() (id int32, meta int16) {
	return 174, 0
}

// EncodeBlock ...
func (PackedIce) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:packed_ice", nil
}

// Hash ...
func (PackedIce) Hash() uint64 {
	return hashPackedIce
}
