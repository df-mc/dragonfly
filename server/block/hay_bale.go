package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// HayBale is a decorative, flammable block that can also be used to
// feed horses, breed llamas, reduce fall damage, and extend campfire smokes.
type HayBale struct {
	solid

	// Axis is the axis which the hay bale block faces.
	Axis cube.Axis

	// Deprecated it has no use, however it must be implemented for it to work.
	Deprecated int32
}

// UseOnBlock ...
func (h HayBale) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, h)
	if !used {
		return
	}
	h.Axis = face.Axis()

	place(w, pos, h, user, ctx)
	return placed(ctx)
}

// FlammabilityInfo ...
func (HayBale) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 20, false)
}

// BreakInfo ...
func (h HayBale) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, hoeEffective, oneOf(h))
}

// EncodeItem ...
func (HayBale) EncodeItem() (name string, meta int16) {
	return "minecraft:hay_block", 0
}

// EncodeBlock ...
func (h HayBale) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:hay_block", map[string]interface{}{"pillar_axis": h.Axis.String(), "deprecated": h.Deprecated}
}

// allHayBales ...
func allHayBales() (haybale []world.Block) {
	var i int32
	for i = 0; i < 4; i++ {
		for _, a := range cube.Axes() {
			haybale = append(haybale, HayBale{Axis: a, Deprecated: i})
		}
	}
	return
}
