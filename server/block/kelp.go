package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// Kelp is an underwater block which can grow on top of solids underwater.
type Kelp struct {
	empty
	transparent
	sourceWaterDisplacer

	// Age is the age of the kelp block which can be 0-25. If age is 25, kelp won't grow any further.
	Age int
}

// SmeltInfo ...
func (k Kelp) SmeltInfo() item.SmeltInfo {
	return newFoodSmeltInfo(item.NewStack(item.DriedKelp{}, 1), 0.1)
}

// BoneMeal ...
func (k Kelp) BoneMeal(pos cube.Pos, tx *world.Tx) bool {
	for y := pos.Y(); y <= tx.Range()[1]; y++ {
		currentPos := cube.Pos{pos.X(), y, pos.Z()}
		block := tx.Block(currentPos)
		if kelp, ok := block.(Kelp); ok {
			if kelp.Age == 25 {
				break
			}
			continue
		}
		if water, ok := block.(Water); ok && water.Depth == 8 {
			tx.SetBlock(currentPos, Kelp{Age: k.Age + 1}, nil)
			return true
		}
		break
	}
	return false
}

// BreakInfo ...
func (k Kelp) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(k))
}

// CompostChance ...
func (Kelp) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (Kelp) EncodeItem() (name string, meta int16) {
	return "minecraft:kelp", 0
}

// EncodeBlock ...
func (k Kelp) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:kelp", map[string]any{"kelp_age": int32(k.Age)}
}

// SideClosed will always return false since kelp doesn't close any side.
func (Kelp) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// withRandomAge returns a new Kelp block with its age value randomized between 0 and 24.
func (k Kelp) withRandomAge() Kelp {
	k.Age = rand.IntN(25)
	return k
}

// UseOnBlock ...
func (k Kelp) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, k)
	if !used {
		return
	}

	below := pos.Side(cube.FaceDown)
	belowBlock := tx.Block(below)
	if _, kelp := belowBlock.(Kelp); !kelp {
		if !belowBlock.Model().FaceSolid(below, cube.FaceUp, tx) {
			return false
		}
	}

	liquid, ok := tx.Liquid(pos)
	if !ok {
		return false
	} else if _, ok := liquid.(Water); !ok || liquid.LiquidDepth() < 8 {
		return false
	}

	// When first placed, kelp gets a random age between 0 and 24.
	place(tx, pos, k.withRandomAge(), user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (k Kelp) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *world.Tx) {
	if _, ok := tx.Liquid(pos); !ok {
		breakBlock(k, pos, tx)
		return
	}
	if changedNeighbour[1]-1 == pos.Y() {
		// When a kelp block is broken above, the kelp block underneath it gets a new random age.
		tx.SetBlock(pos, k.withRandomAge(), nil)
	}

	below := pos.Side(cube.FaceDown)
	belowBlock := tx.Block(below)
	if _, kelp := belowBlock.(Kelp); !kelp {
		if !belowBlock.Model().FaceSolid(below, cube.FaceUp, tx) {
			breakBlock(k, pos, tx)
		}
	}
}

// RandomTick ...
func (k Kelp) RandomTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	// Every random tick, there's a 14% chance for Kelp to grow if its age is below 25.
	if r.IntN(100) < 15 && k.Age < 25 {
		abovePos := pos.Side(cube.FaceUp)

		liquid, ok := tx.Liquid(abovePos)

		// For kelp to grow, there must be only water above.
		if !ok {
			return
		} else if _, ok := liquid.(Water); ok {
			switch tx.Block(abovePos).(type) {
			case Air, Water:
				tx.SetBlock(abovePos, Kelp{Age: k.Age + 1}, nil)
				if liquid.LiquidDepth() < 8 {
					// When kelp grows into a water block, the water block becomes a source block.
					tx.SetLiquid(abovePos, Water{Still: true, Depth: 8, Falling: false})
				}
			}
		}
	}
}

// allKelp returns all possible states of a kelp block.
func allKelp() (b []world.Block) {
	for i := 0; i < 26; i++ {
		b = append(b, Kelp{Age: i})
	}
	return
}
