package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
)

// Kelp is an underwater block which can grow on top of solids underwater.
type Kelp struct {
	noNBT
	empty
	transparent

	// Age is the age of the kelp block which can be 0-25. If age is 25, kelp won't grow any further.
	Age int
}

// Bonemeal ...
func (k Kelp) Bonemeal(pos world.BlockPos, w *world.World) bool {
	for y := pos.Y(); y < 255; y++ {
		currentPos := world.BlockPos{pos.X(), y, pos.Z()}
		block := w.Block(currentPos)
		if kelp, ok := block.(Kelp); ok {
			if kelp.Age == 25 {
				break
			}
			continue
		}
		if water, ok := block.(Water); ok && water.Depth == 8 {
			w.PlaceBlock(currentPos, Kelp{Age: k.Age + 1})
			return true
		}
		break
	}
	return false
}

// BreakInfo ...
func (k Kelp) BreakInfo() BreakInfo {
	// Kelp can be instantly destroyed.
	return BreakInfo{
		Hardness:    0.0,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(k, 1)),
	}
}

// EncodeItem ...
func (Kelp) EncodeItem() (id int32, meta int16) {
	return 335, 0
}

// EncodeBlock ...
func (k Kelp) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:kelp", map[string]interface{}{"kelp_age": int32(k.Age)}
}

// Hash ...
func (k Kelp) Hash() uint64 {
	return hashKelp | (uint64(k.Age) << 32)
}

// CanDisplace will return true if the liquid is Water, since kelp can waterlog.
func (Kelp) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed will always return false since kelp doesn't close any side.
func (Kelp) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
}

// withRandomAge returns a new Kelp block with its age value randomized between 0 and 24.
func (k Kelp) withRandomAge() Kelp {
	k.Age = rand.Intn(25)
	return k
}

// UseOnBlock ...
func (k Kelp) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, k)
	if !used {
		return
	}

	below := pos.Add(world.BlockPos{0, -1})
	belowBlock := w.Block(below)
	if _, kelp := belowBlock.(Kelp); !kelp {
		if !belowBlock.Model().FaceSolid(below, world.FaceUp, w) {
			return false
		}
	}

	liquid, ok := w.Liquid(pos)
	if !ok {
		return false
	} else if _, ok := liquid.(Water); !ok || liquid.LiquidDepth() < 8 {
		return false
	}

	// When first placed, kelp gets a random age between 0 and 24.
	place(w, pos, k.withRandomAge(), user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (k Kelp) NeighbourUpdateTick(pos, changed world.BlockPos, w *world.World) {
	if _, ok := w.Liquid(pos); !ok {
		w.BreakBlock(pos)
		return
	}
	if changed.Y()-1 == pos.Y() {
		// When a kelp block is broken above, the kelp block underneath it gets a new random age.
		w.PlaceBlock(pos, k.withRandomAge())
	}

	below := pos.Add(world.BlockPos{0, -1})
	belowBlock := w.Block(below)
	if _, kelp := belowBlock.(Kelp); !kelp {
		if !belowBlock.Model().FaceSolid(below, world.FaceUp, w) {
			w.BreakBlock(pos)
		}
	}
}

// RandomTick ...
func (k Kelp) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	// Every random tick, there's a 14% chance for Kelp to grow if its age is below 25.
	if r.Intn(100) < 15 && k.Age < 25 {
		abovePos := pos.Add(world.BlockPos{0, 1})

		liquid, ok := w.Liquid(abovePos)

		// For kelp to grow, there must be only water above.
		if !ok {
			return
		} else if _, ok := liquid.(Water); ok {
			switch w.Block(abovePos).(type) {
			case Air, Water:
				w.PlaceBlock(abovePos, Kelp{Age: k.Age + 1})
				if liquid.LiquidDepth() < 8 {
					// When kelp grows into a water block, the water block becomes a source block.
					w.SetLiquid(abovePos, Water{Still: true, Depth: 8, Falling: false})
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
