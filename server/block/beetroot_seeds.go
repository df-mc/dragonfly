package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// BeetrootSeeds are a crop that can be harvested to craft soup or red dye.
type BeetrootSeeds struct {
	crop
}

// SameCrop ...
func (BeetrootSeeds) SameCrop(c Crop) bool {
	_, ok := c.(BeetrootSeeds)
	return ok
}

// BoneMeal ...
func (b BeetrootSeeds) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if b.Growth == 7 {
		return false
	}
	if rand.Float64() < 0.75 {
		b.Growth++
		tx.SetBlock(pos, b, nil)
		return true
	}
	return false
}

// UseOnBlock ...
func (b BeetrootSeeds) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}

	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		return false
	}

	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (b BeetrootSeeds) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, cropSeedDrops(b, item.Beetroot{}, b.Growth))
}

// CompostChance ...
func (BeetrootSeeds) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (b BeetrootSeeds) EncodeItem() (name string, meta int16) {
	return "minecraft:beetroot_seeds", 0
}

// RandomTick ...
func (b BeetrootSeeds) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if tx.Light(pos) < 8 {
		breakBlock(b, pos, tx)
	} else if b.Growth < 7 && r.IntN(3) > 0 && r.Float64() <= b.CalculateGrowthChance(pos, tx) {
		b.Growth++
		tx.SetBlock(pos, b, nil)
	}
}

// EncodeBlock ...
func (b BeetrootSeeds) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:beetroot", map[string]any{"growth": int32(b.Growth)}
}

// allBeetroot ...
func allBeetroot() (beetroot []world.Block) {
	for i := 0; i <= 7; i++ {
		beetroot = append(beetroot, BeetrootSeeds{crop: crop{Growth: i}})
	}
	return
}
