package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Wheat is a crop that can be harvested to craft bread, cake, & cookies.
type Wheat struct {
	noNBT
	transparent
	empty

	// Growth is the current stage of growth. Max is 7.
	Growth int
}

// NeighbourUpdateTick ...
func (s Wheat) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		w.BreakBlock(pos)
	}
}

// UseOnBlock ...
func (s Wheat) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, s)
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

// HasLiquidDrops...
func (s Wheat) HasLiquidDrops() bool {
	return true
}

// RandomTick ...
func (s Wheat) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if s.Growth < 7 && r.Intn(2) == 0 {
		s.Growth++
		w.PlaceBlock(pos, s)
	}
}

// GrowthStage ...
func (s Wheat) GrowthStage() int {
	return s.Growth
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
		wheat = append(wheat, Wheat{Growth: i})
	}
	return
}
