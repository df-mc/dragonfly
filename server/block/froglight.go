package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Froglight is a luminous natural block that can be obtained if a frog eats a tiny magma cube.
type Froglight struct {
	solid

	// Type is the type of froglight.
	Type FroglightType
	// Axis is the axis which the froglight block faces.
	Axis cube.Axis
}

// LightEmissionLevel ...
func (f Froglight) LightEmissionLevel() uint8 {
	return 15
}

// UseOnBlock ...
func (f Froglight) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, f)
	if !used {
		return
	}
	f.Axis = face.Axis()

	place(w, pos, f, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (f Froglight) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, oneOf(f))
}

// EncodeItem ...
func (f Froglight) EncodeItem() (name string, meta int16) {
	return "minecraft:" + f.Type.String() + "_froglight", 0
}

// EncodeBlock ...
func (f Froglight) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + f.Type.String() + "_froglight", map[string]any{"axis": f.Axis.String()}
}

// allFrogLight ...
func allFroglight() (froglight []world.Block) {
	for _, axis := range cube.Axes() {
		for _, t := range FroglightTypes() {
			froglight = append(froglight, Froglight{Type: t, Axis: axis})
		}
	}
	return
}
