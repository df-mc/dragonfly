package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// Potato is a crop that can be consumed raw or cooked to make baked potatoes.
type Potato struct {
	crop
}

// SameCrop ...
func (Potato) SameCrop(c Crop) bool {
	_, ok := c.(Potato)
	return ok
}

// AlwaysConsumable ...
func (p Potato) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (p Potato) ConsumeDuration() time.Duration {
	return item.DefaultConsumeDuration
}

// Consume ...
func (p Potato) Consume(_ *world.World, c item.Consumer) item.Stack {
	c.Saturate(1, 0.6)
	return item.Stack{}
}

// BoneMeal ...
func (p Potato) BoneMeal(pos cube.Pos, w *world.World) bool {
	if p.Growth == 7 {
		return false
	}
	p.Growth = min(p.Growth+rand.Intn(4)+2, 7)
	w.PlaceBlock(pos, p)
	return true
}

// UseOnBlock ...
func (p Potato) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, p)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		return false
	}

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (p Potato) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops: func(t tool.Tool) []item.Stack {
			if rand.Float64() < 0.02 {
				return []item.Stack{item.NewStack(p, rand.Intn(5)+1), item.NewStack(item.PoisonousPotato{}, 1)}
			}
			return []item.Stack{item.NewStack(p, rand.Intn(5)+1)}
		},
	}
}

// EncodeItem ...
func (p Potato) EncodeItem() (id int32, meta int16) {
	return 392, 0
}

// RandomTick ...
func (p Potato) RandomTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if w.Light(pos) < 8 {
		w.BreakBlock(pos)
	} else if p.Growth < 7 && rand.Float64() <= p.CalculateGrowthChance(pos, w) {
		p.Growth++
		w.PlaceBlock(pos, p)
	}
}

// EncodeBlock ...
func (p Potato) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:potatoes", map[string]interface{}{"growth": int32(p.Growth)}
}

// Hash ...
func (p Potato) Hash() uint64 {
	return hashPotato | (uint64(p.Growth) << 32)
}

// allPotato ...
func allPotato() (potato []canEncode) {
	for i := 0; i <= 7; i++ {
		potato = append(potato, Potato{crop{Growth: i}})
	}
	return
}
