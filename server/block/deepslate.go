package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Deepslate is similar to stone but is naturally found deep underground around Y0 and below, and is harder to break.
type Deepslate struct {
	solid
	bassDrum

	// Type is the type of deepslate of the block.
	Type DeepslateType
	// Axis is the axis which the deepslate faces.
	Axis cube.Axis
}

// BreakInfo ...
func (d Deepslate) BreakInfo() BreakInfo {
	if d.Type == NormalDeepslate() {
		return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(Deepslate{Type: CobbledDeepslate()}, d)).withBlastResistance(30)
	}
	return newBreakInfo(3.5, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBlastResistance(30)
}

// SmeltInfo ...
func (d Deepslate) SmeltInfo() item.SmeltInfo {
	if d.Type == CobbledDeepslate() {
		return newSmeltInfo(item.NewStack(Deepslate{}, 1), 0.1)
	}
	return item.SmeltInfo{}
}

// UseOnBlock ...
func (d Deepslate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, d)
	if !used {
		return
	}
	if d.Type == NormalDeepslate() {
		d.Axis = face.Axis()
	}

	place(tx, pos, d, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (d Deepslate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + d.Type.String(), 0
}

// EncodeBlock ...
func (d Deepslate) EncodeBlock() (string, map[string]any) {
	if d.Type == NormalDeepslate() {
		return "minecraft:deepslate", map[string]any{"pillar_axis": d.Axis.String()}
	}
	return "minecraft:" + d.Type.String(), nil
}

// allDeepslate returns a list of all deepslate block variants.
func allDeepslate() (s []world.Block) {
	for _, t := range DeepslateTypes() {
		axes := []cube.Axis{0}
		if t == NormalDeepslate() {
			axes = cube.Axes()
		}
		for _, axis := range axes {
			s = append(s, Deepslate{Type: t, Axis: axis})
		}
	}
	return
}
