package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// NetherWart is a fungus found in the Nether that is vital in the creation of potions.
type NetherWart struct {
	noNBT
	transparent
	empty

	// Age is the age of the nether wart block. 3 is fully grown.
	Age int
}

// HasLiquidDrops ...
func (n NetherWart) HasLiquidDrops() bool {
	return true
}

// RandomTick ...
func (n NetherWart) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	if n.Age < 3 && r.Float64() < 0.1 {
		n.Age++
		w.PlaceBlock(pos, n)
	}
}

// UseOnBlock ...
func (n NetherWart) UseOnBlock(pos world.BlockPos, face world.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, n)
	if !used {
		return false
	}
	if _, ok := w.Block(pos.Side(world.FaceDown)).(SoulSand); !ok {
		return false
	}

	place(w, pos, n, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (n NetherWart) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Side(world.FaceDown)).(SoulSand); !ok {
		w.BreakBlockWithoutParticles(pos)
	}
}

// BreakInfo ...
func (n NetherWart) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if n.Age == 3 {
				return []item.Stack{item.NewStack(n, rand.Intn(3)+2)}
			}
			return []item.Stack{item.NewStack(n, 1)}
		},
	}
}

// EncodeItem ...
func (NetherWart) EncodeItem() (id int32, meta int16) {
	return 372, 0
}

// EncodeBlock ...
func (n NetherWart) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:nether_wart", map[string]interface{}{"age": int32(n.Age)}
}

// Hash ...
func (n NetherWart) Hash() uint64 {
	return hashNetherWart | (uint64(n.Age) << 32)
}

// allNetherWart ...
func allNetherWart() (wart []canEncode) {
	for i := 0; i < 4; i++ {
		wart = append(wart, NetherWart{Age: i})
	}
	return
}
