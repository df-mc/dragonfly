package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// Kelp is an underwater block which can grow on top of solids underwater.
type Kelp struct {
	noNBT
	empty
	transparent

	// Age is the age of the kelp block which can be 0-15. If age is 15, kelp won't grow any further.
	Age int
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
	return "minecraft:kelp", map[string]interface{}{"age": int32(k.Age)}
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

// withRandomAge returns a new Kelp block with its age value randomized between 0 and 14.
func (k Kelp) withRandomAge() Kelp {
	// In Java Edition, Kelp's age value can be up to 25, but MCPE limits it to 15.
	k.Age = rand.Intn(14)
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

	// When first placed, kelp gets a random age between 0 and 14 in MCBE.
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
			w.ScheduleBlockUpdate(pos, time.Second/20)
		}
	}
}

// ScheduledTick ...
func (Kelp) ScheduledTick(pos world.BlockPos, w *world.World) {
	if _, ok := w.Liquid(pos); !ok {
		w.BreakBlock(pos)
		return
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
	if r.Intn(100) < 15 && k.Age < 15 {
		// Every random tick, there's a 14% chance for Kelp to grow if its age is below 15.
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
	w.ScheduleBlockUpdate(pos, time.Second/20)
}

// allKelp returns all possible states of a kelp block.
func allKelp() (b []world.Block) {
	for i := 0; i < 16; i++ {
		b = append(b, Kelp{Age: i})
	}
	return
}
