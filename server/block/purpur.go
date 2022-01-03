package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Purpur is a decorative block that is naturally generated in End cities and End ships.
type Purpur struct {
	solid
	bassDrum

	// Type is the type of purpur of the block.
	Type PurpurType

	// Axis is the axis which the purpur pillar block faces.
	Axis cube.Axis
}

// UseOnBlock ...
func (p Purpur) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}
	if p.Type == PillarPurpur() {
		p.Axis = face.Axis()
	}

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (p Purpur) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(p))
}

// EncodeItem ...
func (p Purpur) EncodeItem() (name string, meta int16) {
	return "minecraft:purpur_block", int16(p.Type.Uint8())
}

// EncodeBlock ...
func (p Purpur) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:purpur_block", map[string]interface{}{"chisel_type": p.Type.String(), "pillar_axis": p.Axis.String()}
}

// allPurpurs ...
func allPurpurs() (purpur []world.Block) {
	for _, p := range PurpurTypes() {
		if p == PillarPurpur() {
			for _, axis := range cube.Axes() {
				purpur = append(purpur, Purpur{Type: p, Axis: axis})
			}
			continue
		}
		purpur = append(purpur, Purpur{Type: p})
	}
	return
}
