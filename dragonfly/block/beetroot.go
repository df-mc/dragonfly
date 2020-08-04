package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Beetroot is a crop that can be harvested to craft soup or red dye.
type Beetroot struct {
	crop
}

// Bonemeal ...
func (b Beetroot) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if b.Growth == 7 {
		return false
	}
	if rand.Intn(100) < 75 {
		b.Growth++
		w.PlaceBlock(pos, b)
		return true
	}
	return false
}

// UseOnBlock ...
func (b Beetroot) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
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
func (b Beetroot) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if b.Growth < 7 {
				return []item.Stack{item.NewStack(b, 1)}
			}
			//TODO: Beetroot item
			return []item.Stack{item.NewStack(b, rand.Intn(4))}
		},
	}
}

// EncodeItem ...
func (b Beetroot) EncodeItem() (id int32, meta int16) {
	return 458, 0
}

// RandomTick ...
func (b Beetroot) RandomTick(pos world.BlockPos, w *world.World, _ *rand.Rand) {
	if b.Growth < 7 && rand.Float64() <= b.CalculateGrowthChance(pos, w) {
		b.Growth++
		w.PlaceBlock(pos, b)
	}
}

// EncodeBlock ...
func (b Beetroot) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:beetroot", map[string]interface{}{"growth": int32(b.Growth)}
}

// Hash ...
func (b Beetroot) Hash() uint64 {
	return hashBeetroot | (uint64(b.Growth) << 32)
}

// allBeetroot ...
func allBeetroot() (beetroot []world.Block) {
	for i := 0; i <= 7; i++ {
		beetroot = append(beetroot, Beetroot{crop{Growth: i}})
	}
	return
}
