package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SporeBlossom is a decorative block.
type SporeBlossom struct {
	empty
	transparent
}

func (s SporeBlossom) HasLiquidDrops() bool {
	return true
}

func (s SporeBlossom) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !tx.Block(pos.Side(cube.FaceUp)).Model().FaceSolid(pos.Side(cube.FaceUp), cube.FaceDown, tx) {
		breakBlock(s, pos, tx)
	}
}

func (s SporeBlossom) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, s)
	if !used {
		return
	}
	if !tx.Block(pos.Side(cube.FaceUp)).Model().FaceSolid(pos.Side(cube.FaceUp), cube.FaceDown, tx) {
		return
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

func (s SporeBlossom) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(s))
}

func (SporeBlossom) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(15, 100, true)
}

func (SporeBlossom) CompostChance() float64 {
	return 0.65
}

func (s SporeBlossom) EncodeItem() (name string, meta int16) {
	return "minecraft:spore_blossom", 0
}

func (s SporeBlossom) EncodeBlock() (string, map[string]any) {
	return "minecraft:spore_blossom", nil
}
