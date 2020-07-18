package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// Glass is a decorative, fully transparent solid block that can be dyed into stained glass.
type Glass struct{}

// BreakInfo ...
func (g Glass) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0.3,
		Drops:    simpleDrops(),
		Harvestable: func(t tool.Tool) bool {
			return true
		},
		Effective: nothingEffective,
	}
}

// EncodeItem ...
func (g Glass) EncodeItem() (id int32, meta int16) {
	return 20, 0
}

// EncodeBlock ...
func (g Glass) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:glass", nil
}

// Hash ...
func (Glass) Hash() uint64 {
	return hashGlass
}
