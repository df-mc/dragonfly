package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// WheatSeeds are a crop that can be harvested to craft bread, cake, & cookies.
type WheatSeeds struct {
	crop
}

// SameCrop ...
func (WheatSeeds) SameCrop(c Crop) bool {
	_, ok := c.(WheatSeeds)
	return ok
}

// BoneMeal ...
func (s WheatSeeds) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if s.Growth == 7 {
		return false
	}
	s.Growth = min(s.Growth+rand.IntN(4)+2, 7)
	tx.SetBlock(pos, s, nil)
	return true
}

// UseOnBlock ...
func (s WheatSeeds) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, s)
	if !used {
		return false
	}

	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		return false
	}

	place(tx, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s WheatSeeds) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, cropSeedDrops(s, item.Wheat{}, s.Growth))
}

// CompostChance ...
func (WheatSeeds) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (s WheatSeeds) EncodeItem() (name string, meta int16) {
	return "minecraft:wheat_seeds", 0
}

// RandomTick ...
func (s WheatSeeds) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if tx.Light(pos) < 8 {
		breakBlock(s, pos, tx)
	} else if s.Growth < 7 && r.Float64() <= s.CalculateGrowthChance(pos, tx) {
		s.Growth++
		tx.SetBlock(pos, s, nil)
	}
}

// EncodeBlock ...
func (s WheatSeeds) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:wheat", map[string]any{"growth": int32(s.Growth)}
}

// allWheat ...
func allWheat() (wheat []world.Block) {
	for i := 0; i <= 7; i++ {
		wheat = append(wheat, WheatSeeds{crop{Growth: i}})
	}
	return
}
