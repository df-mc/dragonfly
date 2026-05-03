package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// InfestedDeepslate is a block that hides a silverfish. It looks identical to deepslate.
type InfestedDeepslate struct {
	solid
	bassDrum

	// Axis is the axis which the deepslate faces.
	Axis cube.Axis
}

// BreakInfo ...
func (i InfestedDeepslate) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, silkTouchOnlyDrop(Deepslate{Axis: i.Axis})).withBlastResistance(0.75)
}

// UseOnBlock ...
func (i InfestedDeepslate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, i)
	if !used {
		return
	}
	i.Axis = face.Axis()

	place(tx, pos, i, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (i InfestedDeepslate) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_deepslate", 0
}

// EncodeBlock ...
func (i InfestedDeepslate) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_deepslate", map[string]any{"pillar_axis": i.Axis.String()}
}

// allInfestedDeepslate ...
func allInfestedDeepslate() (s []world.Block) {
	for _, axis := range cube.Axes() {
		s = append(s, InfestedDeepslate{Axis: axis})
	}
	return
}
