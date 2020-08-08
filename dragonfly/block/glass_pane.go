package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// GlassPane is a transparent block that can be used as a more efficient alternative to glass blocks.
type GlassPane struct {
	noNBT
	transparent
	thin
}

// CanDisplace ...
func (p GlassPane) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (p GlassPane) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
}

// BreakInfo ...
func (p GlassPane) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 0.3,
		Harvestable: func(t tool.Tool) bool {
			return true // TODO(lhochbaum): Glass panes can be silk touched, implement silk touch.
		},
		Effective: nothingEffective,
		Drops:     simpleDrops(),
	}
}

// EncodeItem ...
func (p GlassPane) EncodeItem() (id int32, meta int16) {
	return 102, meta
}

// EncodeBlock ...
func (p GlassPane) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:glass_pane", map[string]interface{}{}
}

// Hash ...
func (p GlassPane) Hash() uint64 {
	return hashGlassPane
}
