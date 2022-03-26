package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
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
func (b BeetrootSeeds) BoneMeal(pos cube.Pos, w *world.World) bool {
	if b.Growth == 7 {
		return false
	}
	if rand.Float64() < 0.75 {
		b.Growth++
		w.SetBlock(pos, b, nil)
		return true
	}
	return false
}

// UseOnBlock ...
func (b BeetrootSeeds) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, b)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		return false
	}

	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (b BeetrootSeeds) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(item.Tool, []item.Enchantment) []item.Stack {
		if b.Growth < 7 {
			return []item.Stack{item.NewStack(b, 1)}
		}
		return []item.Stack{item.NewStack(item.Beetroot{}, 1), item.NewStack(b, rand.Intn(4)+1)}
	})
}

// EncodeItem ...
func (b BeetrootSeeds) EncodeItem() (name string, meta int16) {
	return "minecraft:beetroot_seeds", 0
}

// RandomTick ...
func (b BeetrootSeeds) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if w.Light(pos) < 8 {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: b})
	} else if b.Growth < 7 && r.Intn(3) > 0 && r.Float64() <= b.CalculateGrowthChance(pos, w) {
		b.Growth++
		w.SetBlock(pos, b, nil)
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
