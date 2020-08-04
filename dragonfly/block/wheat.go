package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
)

// Wheat is a crop that can be harvested to craft bread, cake, & cookies.
type Wheat struct {
	crop
}

// Bonemeal ...
func (s Wheat) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if s.Growth == 7 {
		return false
	}
	s.Growth = int(math.Min(float64(s.Growth+rand.Intn(4)+2), 7))
	w.PlaceBlock(pos, s)
	return true
}

// NeighbourUpdateTick ...
func (s Wheat) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		w.BreakBlock(pos)
	}
}

// UseOnBlock ...
func (s Wheat) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
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
func (s Wheat) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if s.Growth < 7 {
				return []item.Stack{item.NewStack(s, 1)}
			}
			return []item.Stack{item.NewStack(item.Wheat{}, 1), item.NewStack(s, rand.Intn(4))}
		},
	}
}

// EncodeItem ...
func (s Wheat) EncodeItem() (id int32, meta int16) {
	return 295, 0
}

// HasLiquidDrops ...
func (s Wheat) HasLiquidDrops() bool {
	return true
}

// RandomTick ...
func (s Wheat) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if s.Growth < 7 && rand.Float64() <= s.CalculateGrowthChance(s, pos, w) {
		s.Growth++
		w.PlaceBlock(pos, s)
	}
}

// EncodeBlock ...
func (s Wheat) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:wheat", map[string]interface{}{"growth": int32(s.Growth)}
}

// Hash ...
func (s Wheat) Hash() uint64 {
	return hashWheat | (uint64(s.Growth) << 32)
}

// allWheat ...
func allWheat() (wheat []world.Block) {
	for i := 0; i <= 7; i++ {
		wheat = append(wheat, Wheat{crop{Growth: i}})
	}
	return
}
