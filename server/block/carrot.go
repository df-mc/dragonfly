package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// Carrot is a crop that can be consumed raw.
type Carrot struct {
	crop
}

func (Carrot) SameCrop(c Crop) bool {
	_, ok := c.(Carrot)
	return ok
}

func (c Carrot) AlwaysConsumable() bool {
	return false
}

func (c Carrot) ConsumeDuration() time.Duration {
	return item.DefaultConsumeDuration
}

func (c Carrot) Consume(_ *world.Tx, co item.Consumer) item.Stack {
	co.Saturate(3, 3.6)
	return item.Stack{}
}

func (c Carrot) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	if c.Growth == 7 {
		return false
	}
	c.Growth = min(c.Growth+rand.IntN(4)+2, 7)
	tx.SetBlock(pos, c, nil)
	return true
}

func (c Carrot) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, c)
	if !used {
		return false
	}

	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		return false
	}

	place(tx, pos, c, user, ctx)
	return placed(ctx)
}

func (c Carrot) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(item.Tool, []item.Enchantment) []item.Stack {
		if c.Growth < 7 {
			return []item.Stack{item.NewStack(c, 1)}
		}
		return []item.Stack{item.NewStack(c, rand.IntN(4)+2)}
	})
}

func (Carrot) CompostChance() float64 {
	return 0.65
}

func (c Carrot) EncodeItem() (name string, meta int16) {
	return "minecraft:carrot", 0
}

func (c Carrot) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if tx.Light(pos) < 8 {
		breakBlock(c, pos, tx)
	} else if c.Growth < 7 && r.Float64() <= c.CalculateGrowthChance(pos, tx) {
		c.Growth++
		tx.SetBlock(pos, c, nil)
	}
}

func (c Carrot) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:carrots", map[string]any{"growth": int32(c.Growth)}
}

func allCarrots() (carrots []world.Block) {
	for growth := 0; growth < 8; growth++ {
		carrots = append(carrots, Carrot{crop{Growth: growth}})
	}
	return
}
