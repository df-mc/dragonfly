package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// GlassPane is a transparent block that can be used as a more efficient alternative to glass blocks.
type GlassPane struct {}

// BreakInfo ...
func (p GlassPane) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.3,
		Harvestable: func(t tool.Tool) bool {
			return true // TODO(lhochbaum): Glass panes can be silk touched, implement silk touch.
		},
		Effective:   nothingEffective,
		Drops:       simpleDrops(),
	}
}

// EncodeItem ...
func (p GlassPane) EncodeItem() (id int32, meta int16) {
	return 102, meta
}

// EncodeBlock ...
func (g GlassPane) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:glass_pane", map[string]interface{}{}
}

// TODO(lhochbaum): Adjust the bounding box.