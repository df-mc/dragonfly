package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type (
	// Purpur is a decorative block that is naturally generated in End cities and End ships.
	Purpur struct {
		solid
		bassDrum
	}
	// PurpurPillar is a variant of Purpur that can be rotated.
	PurpurPillar struct {
		solid
		bassDrum

		// Axis is the axis which the purpur pillar block faces.
		Axis cube.Axis
	}
)

func (p Purpur) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(p)).withBlastResistance(30)
}

func (p Purpur) EncodeItem() (name string, meta int16) {
	return "minecraft:purpur_block", 0
}

func (p Purpur) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:purpur_block", map[string]interface{}{"pillar_axis": "y"}
}

func (p PurpurPillar) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, p)
	if !used {
		return
	}
	p.Axis = face.Axis()

	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

func (p PurpurPillar) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(p)).withBlastResistance(30)
}

func (p PurpurPillar) EncodeItem() (name string, meta int16) {
	return "minecraft:purpur_pillar", 0
}

func (p PurpurPillar) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:purpur_pillar", map[string]interface{}{"pillar_axis": p.Axis.String()}
}

func allPurpurs() (purpur []world.Block) {
	purpur = append(purpur, Purpur{})
	for _, axis := range cube.Axes() {
		purpur = append(purpur, PurpurPillar{Axis: axis})
	}
	return
}
