package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// WheatSeeds are a crop that can be harvested to craft bread, cake, & cookies.
type WheatSeeds struct {
	crop
}

// Bonemeal ...
func (s WheatSeeds) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if s.Growth == 7 {
		return false
	}
	s.Growth = min(s.Growth+rand.Intn(4)+2, 7)
	w.PlaceBlock(pos, s)
	return true
}

// UseOnBlock ...
func (s WheatSeeds) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, s)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		return false
	}

	place(w, pos, s, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (s WheatSeeds) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if s.Growth < 7 {
				return []item.Stack{item.NewStack(s, 1)}
			}
			return []item.Stack{item.NewStack(item.Wheat{}, 1), item.NewStack(s, rand.Intn(4)+1)}
		},
	}
}

// EncodeItem ...
func (s WheatSeeds) EncodeItem() (id int32, meta int16) {
	return 295, 0
}

// RandomTick ...
func (s WheatSeeds) RandomTick(pos world.BlockPos, w *world.World, _ *rand.Rand) {
	if w.Light(pos) < 8 {
		w.BreakBlock(pos)
	} else if s.Growth < 7 && rand.Float64() <= s.CalculateGrowthChance(pos, w) {
		s.Growth++
		w.PlaceBlock(pos, s)
	}
}

// EncodeBlock ...
func (s WheatSeeds) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:wheat", map[string]interface{}{"growth": int32(s.Growth)}
}

// Hash ...
func (s WheatSeeds) Hash() uint64 {
	return hashWheat | (uint64(s.Growth) << 32)
}

// allWheat ...
func allWheat() (wheat []world.Block) {
	for i := 0; i <= 7; i++ {
		wheat = append(wheat, WheatSeeds{crop{Growth: i}})
	}
	return
}
