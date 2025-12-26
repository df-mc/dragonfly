package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Fern is a transparent plant block which can be used to obtain seeds and as decoration.
type Fern struct {
	replaceable
	transparent
	empty
}

// FlammabilityInfo ...
func (g Fern) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (g Fern) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, grassDrops(g))
}

// BoneMeal attempts to affect the block using a bone meal item.
func (g Fern) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	upper := DoubleTallGrass{Type: FernDoubleTallGrass(), UpperPart: true}
	if replaceableWith(tx, pos.Side(cube.FaceUp), upper) {
		tx.SetBlock(pos, DoubleTallGrass{Type: FernDoubleTallGrass()}, nil)
		tx.SetBlock(pos.Side(cube.FaceUp), upper, nil)
		return true
	}
	return false
}

// CompostChance ...
func (g Fern) CompostChance() float64 {
	return 0.3
}

// NeighbourUpdateTick ...
func (g Fern) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(g, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(g, pos, tx)
	}
}

// HasLiquidDrops ...
func (g Fern) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (g Fern) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, g)
	if !used || !supportsVegetation(g, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(tx, pos, g, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (g Fern) EncodeItem() (name string, meta int16) {
	return "minecraft:fern", 0
}

// EncodeBlock ...
func (g Fern) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:fern", nil
}
