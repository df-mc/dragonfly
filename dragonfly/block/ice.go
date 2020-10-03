package block

import "github.com/df-mc/dragonfly/dragonfly/item/tool"

// TODO: Slipperiness, melting and SilkTouch

// Ice is a solid block similar to packed ice.
type Ice struct {
	noNBT
	transparent
	solid
}

// BreakInfo ...
func (i Ice) BreakInfo() BreakInfo {
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
func (Ice) EncodeItem() (id int32, meta int16) {
	return 79, 0
}

// EncodeBlock ...
func (Ice) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:ice", nil
}

// Hash ...
func (Ice) Hash() uint64 {
	return hashIce
}

// LightDiffusionLevel ...
func (i Ice) LightDiffusionLevel() uint8 {
	return 2
}
