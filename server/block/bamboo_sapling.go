package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// BambooSapling ...
type BambooSapling struct {
	empty
	transparent
	bass

	Ready bool
}

var (
	_ item.BoneMealAffected = BambooSapling{}
	_ Flammable             = BambooSapling{}
)

// BoneMeal ...
func (b BambooSapling) BoneMeal(pos cube.Pos, tx *world.Tx) item.BoneMealResult {
	if b.grow(pos, tx) {
		return item.BoneMealResultSmall
	}
	return item.BoneMealResultNone
}

// FlammabilityInfo ...
func (b BambooSapling) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 60, true)
}

// NeighbourUpdateTick ...
func (b BambooSapling) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	down := tx.Block(pos.Side(cube.FaceDown))
	if supportsVegetation(b, down) {
		return
	}
	breakBlock(b, pos, tx)
}

// RandomTick ...
func (b BambooSapling) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if tx.Light(pos) >= 9 && r.IntN(3) == 0 {
		b.grow(pos, tx)
	}
}

// BreakInfo ...
func (b BambooSapling) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, axeEffective, oneOf(Bamboo{}))
}

// HasLiquidDrops ...
func (b BambooSapling) HasLiquidDrops() bool {
	return true
}

// EncodeBlock ...
func (b BambooSapling) EncodeBlock() (string, map[string]any) {
	return "minecraft:bamboo_sapling", map[string]any{"age_bit": boolByte(b.Ready)}
}

// grow ...
func (b BambooSapling) grow(pos cube.Pos, tx *world.Tx) bool {
	if !replaceableWith(tx, pos.Side(cube.FaceUp), b) {
		return false
	}

	tx.SetBlock(pos, Bamboo{}, nil)
	tx.SetBlock(pos.Side(cube.FaceUp), Bamboo{LeafSize: BambooSizeSmallLeaves()}, nil)
	return true
}

// allBambooSaplings ...
func allBambooSaplings() (saplings []world.Block) {
	saplings = append(saplings, BambooSapling{Ready: false})
	saplings = append(saplings, BambooSapling{Ready: true})
	return
}
