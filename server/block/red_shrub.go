package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedShrub is a small decorative plant that grows on dirt. Bone meal spreads it to the blocks around it.
type RedShrub struct {
	replaceable
	transparent
	empty
}

// FlammabilityInfo ...
func (RedShrub) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, false)
}

// BreakInfo ...
func (r RedShrub) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(r))
}

// BoneMeal grows another red shrub on one of the blocks around this one.
func (r RedShrub) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	for _, offset := range rand.Perm(len(shrubSpread)) {
		at := pos.Add(shrubSpread[offset])
		if !replaceableWith(tx, at, r) || !supportsVegetation(r, tx.Block(at.Side(cube.FaceDown))) {
			continue
		}
		tx.SetBlock(at, r, nil)
		return item.BoneMealResultSmall
	}
	return item.BoneMealResultNone
}

// shrubSpread holds the blocks a red shrub spreads to, being those around it on its own level and one above
// and below it.
var shrubSpread = []cube.Pos{
	{1}, {-1}, {0, 0, 1}, {0, 0, -1},
	{1, 1}, {-1, 1}, {0, 1, 1}, {0, 1, -1},
	{1, -1}, {-1, -1}, {0, -1, 1}, {0, -1, -1},
}

// CompostChance ...
func (RedShrub) CompostChance() float64 {
	return 0.3
}

// HasLiquidDrops ...
func (RedShrub) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (r RedShrub) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(r, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(r, pos, tx)
	}
}

// UseOnBlock ...
func (r RedShrub) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used || !supportsVegetation(r, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// EncodeItem ...
func (RedShrub) EncodeItem() (name string, meta int16) {
	return "minecraft:red_shrub", 0
}

// EncodeBlock ...
func (RedShrub) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:red_shrub", nil
}
