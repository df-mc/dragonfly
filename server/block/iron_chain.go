package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// IronChain is a metallic decoration block.
type IronChain struct {
	transparent
	sourceWaterDisplacer

	// Axis is the axis which the chain faces.
	Axis cube.Axis
}

// SideClosed ...
func (IronChain) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// UseOnBlock ...
func (c IronChain) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, c)
	if !used {
		return
	}
	c.Axis = face.Axis()

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c IronChain) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

// EncodeItem ...
func (IronChain) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_chain", 0
}

// EncodeBlock ...
func (c IronChain) EncodeBlock() (string, map[string]any) {
	return "minecraft:iron_chain", map[string]any{"pillar_axis": c.Axis.String()}
}

// Model ...
func (c IronChain) Model() world.BlockModel {
	return model.Chain{Axis: c.Axis}
}

// allIronChains ...
func allIronChains() (chains []world.Block) {
	for _, axis := range cube.Axes() {
		chains = append(chains, IronChain{Axis: axis})
	}
	return
}
