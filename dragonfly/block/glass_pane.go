package block

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// GlassPane is a transparent block that can be used as a more efficient alternative to glass blocks.
type GlassPane struct{}

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

// AABB adjusts bounding box of the glass pane.
func (p GlassPane) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	return calculateThinBounds(pos, w)
}

// calculateThinBounds checks the connections of a thin block in all directions and changes its physics.AABB
// accordingly.
func calculateThinBounds(pos world.BlockPos, w *world.World) []physics.AABB {
	const offset = 0.4375

	boxes := make([]physics.AABB, 0, 5)
	mainBox := physics.NewAABB(mgl64.Vec3{offset, 0, offset}, mgl64.Vec3{1 - offset, 1, 1 - offset})
	thin := w.Block(pos)

	for i := world.Face(2); i < 6; i++ {
		pos := pos.Side(i)
		block := w.Block(pos)
		if partial, isPartiallySolid := block.(PartiallySolid); !isPartiallySolid || partial.FaceSolidTo(pos, i.Opposite(), thin) {
			boxes = append(boxes, mainBox.ExtendTowards(int(i), offset))
		}
	}
	return append(boxes, mainBox)
}
