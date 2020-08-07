package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// Carrot is a crop that can be consumed raw.
type Carrot struct {
	crop
}

// AlwaysConsumable ...
func (c Carrot) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (c Carrot) ConsumeDuration() time.Duration {
	return item.DefaultConsumeDuration
}

// Consume ...
func (c Carrot) Consume(_ *world.World, consumer item.Consumer) item.Stack {
	consumer.Saturate(3, 3.6)
	return item.Stack{}
}

// Bonemeal ...
func (c Carrot) Bonemeal(pos world.BlockPos, w *world.World) bool {
	if c.Growth == 7 {
		return false
	}
	c.Growth = min(c.Growth+rand.Intn(4)+2, 7)
	w.PlaceBlock(pos, c)
	return true
}

// UseOnBlock ...
func (c Carrot) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, c)
	if !used {
		return false
	}

	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		return false
	}

	place(w, pos, c, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (c Carrot) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(c, rand.Intn(5)+1)),
	}
}

// EncodeItem ...
func (c Carrot) EncodeItem() (id int32, meta int16) {
	return 391, 0
}

// RandomTick ...
func (c Carrot) RandomTick(pos world.BlockPos, w *world.World, _ *rand.Rand) {
	if w.Light(pos) < 8 {
		w.BreakBlock(pos)
	} else if c.Growth < 7 && rand.Float64() <= c.CalculateGrowthChance(pos, w) {
		c.Growth++
		w.PlaceBlock(pos, c)
	}
}

// EncodeBlock ...
func (c Carrot) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:carrots", map[string]interface{}{"growth": int32(c.Growth)}
}

// Hash ...
func (c Carrot) Hash() uint64 {
	return hashCarrot | (uint64(c.Growth) << 32)
}

// allCarrot ...
func allCarrot() (carrot []world.Block) {
	for i := 0; i <= 7; i++ {
		carrot = append(carrot, Carrot{crop{Growth: i}})
	}
	return
}
