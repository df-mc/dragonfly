package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// ShortGrass is a transparent plant block which can be used to obtain seeds and as decoration.
type ShortGrass struct {
	replaceable
	transparent
	empty

	Double bool
}

// FlammabilityInfo ...
func (g ShortGrass) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (g ShortGrass) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, grassDrops(g))
}

// BoneMeal attempts to affect the block using a bone meal item.
func (g ShortGrass) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	upper := DoubleTallGrass{Type: NormalDoubleTallGrass(), UpperPart: true}
	if replaceableWith(tx, pos.Side(cube.FaceUp), upper) {
		tx.SetBlock(pos, DoubleTallGrass{Type: NormalDoubleTallGrass()}, nil)
		tx.SetBlock(pos.Side(cube.FaceUp), upper, nil)
		return true
	}
	return false
}

// CompostChance ...
func (g ShortGrass) CompostChance() float64 {
	return 0.65
}

// NeighbourUpdateTick ...
func (g ShortGrass) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(g, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(g, pos, tx)
	}
}

// HasLiquidDrops ...
func (g ShortGrass) HasLiquidDrops() bool {
	return true
}

// UseOnBlock ...
func (g ShortGrass) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, g)
	if !used || !supportsVegetation(g, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(tx, pos, g, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (g ShortGrass) EncodeItem() (name string, meta int16) {
	return "minecraft:short_grass", 0
}

// EncodeBlock ...
func (g ShortGrass) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:short_grass", nil
}
