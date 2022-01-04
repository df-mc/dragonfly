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

	// Pillar specifies if the block is the pillar variant or not.
	Pillar bool

	// Axis is the axis which the purpur pillar block faces.
	Axis cube.Axis
}

// UseOnBlock ...
func (p Purpur) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, p)
	if !used {
		return
	}
	if p.Pillar {
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
	if p.Pillar {
		return "minecraft:purpur_block", 1
	}
	return "minecraft:purpur_block", 0
}

// EncodeBlock ...
func (p Purpur) EncodeBlock() (name string, properties map[string]interface{}) {
	if p.Pillar {
		return "minecraft:purpur_block", map[string]interface{}{"chisel_type": "lines", "pillar_axis": p.Axis.String()}
	}
	return "minecraft:purpur_block", map[string]interface{}{"chisel_type": "default", "pillar_axis": "y"}
}

// allPurpurs ...
func allPurpurs() (purpur []world.Block) {
	purpur = append(purpur, Purpur{})
	for _, axis := range cube.Axes() {
		purpur = append(purpur, Purpur{Pillar: true, Axis: axis})
	}
	return
}
