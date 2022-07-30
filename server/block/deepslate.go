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
	hardness := 3.5
	if d.Type == NormalDeepslate() {
		hardness = 3
	}
	return newBreakInfo(hardness, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBlastResistance(18)
}

// UseOnBlock ...
func (d Deepslate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, d)
	if !used {
		return
	}
	if d.Type == NormalDeepslate() {
		d.Axis = face.Axis()
	}

	place(w, pos, d, user, ctx)
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
		s = append(s, Deepslate{Type: t})
	}
	return
}
