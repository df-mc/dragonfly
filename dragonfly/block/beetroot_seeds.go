package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// BeetrootSeeds are a crop that can be harvested to craft soup or red dye.
type BeetrootSeeds struct {
	crop
}

// Bonemeal ...
func (b BeetrootSeeds) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if b.Growth == 7 {
		return false
	}
	if rand.Float64() < 0.75 {
		b.Growth++
		w.PlaceBlock(pos, b)
		return true
	}
	return false
}

// UseOnBlock ...
func (b BeetrootSeeds) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, b)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		return false
	}

	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (b BeetrootSeeds) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if b.Growth < 7 {
				return []item.Stack{item.NewStack(b, 1)}
			}
			return []item.Stack{item.NewStack(item.Beetroot{}, 1), item.NewStack(b, rand.Intn(4))}
		},
	}
}

// EncodeItem ...
func (b BeetrootSeeds) EncodeItem() (id int32, meta int16) {
	return 458, 0
}

// RandomTick ...
func (b BeetrootSeeds) RandomTick(pos world.BlockPos, w *world.World, _ *rand.Rand) {
	if w.Light(pos) < 8 {
		w.BreakBlock(pos)
	} else if b.Growth < 7 && rand.Intn(3) > 0 && rand.Float64() <= b.CalculateGrowthChance(pos, w) {
		b.Growth++
		w.PlaceBlock(pos, b)
	}
}

// EncodeBlock ...
func (b BeetrootSeeds) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:beetroot", map[string]interface{}{"growth": int32(b.Growth)}
}

// Hash ...
func (b BeetrootSeeds) Hash() uint64 {
	return hashBeetroot | (uint64(b.Growth) << 32)
}

// allBeetroot ...
func allBeetroot() (beetroot []world.Block) {
	for i := 0; i <= 7; i++ {
		beetroot = append(beetroot, BeetrootSeeds{crop{Growth: i}})
	}
	return
}
