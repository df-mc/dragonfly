package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
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
func (b Deepslate) BreakInfo() BreakInfo {
	hardness := 3.5
	if b.Type == NormalDeepslate() {
		hardness = 3
	}
	return newBreakInfo(hardness, pickaxeHarvestable, pickaxeEffective, oneOf(b)).withBlastResistance(18)
}

// EncodeItem ...
func (b Deepslate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Type.String(), 0
}

// EncodeBlock ...
func (b Deepslate) EncodeBlock() (string, map[string]any) {
	if b.Type == NormalDeepslate() {
		return "minecraft:deepslate", map[string]any{"pillar_axis": b.Axis.String()}
	}
	return "minecraft:" + b.Type.String(), nil
}

// allDeepslate returns a list of all deepslate block variants.
func allDeepslate() (s []world.Block) {
	for _, t := range DeepslateTypes() {
		s = append(s, Deepslate{Type: t})
	}
	return
}
