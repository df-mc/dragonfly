package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Chain is a metallic decoration block.
type Chain struct {
	transparent
	sourceWaterDisplacer

	// Axis is the axis which the chain faces.
	Axis cube.Axis
}

func (Chain) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (c Chain) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}
	c.Axis = face.Axis()

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

func (c Chain) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

func (Chain) EncodeItem() (name string, meta int16) {
	return "minecraft:chain", 0
}

func (c Chain) EncodeBlock() (string, map[string]any) {
	return "minecraft:chain", map[string]any{"pillar_axis": c.Axis.String()}
}

func (c Chain) Model() world.BlockModel {
	return model.Chain{Axis: c.Axis}
}

func allChains() (chains []world.Block) {
	for _, axis := range cube.Axes() {
		chains = append(chains, Chain{Axis: axis})
	}
	return
}
