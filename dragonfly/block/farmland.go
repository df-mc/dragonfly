package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"math/rand"
)

// Farmland is a block that crops are grown on. Farmland is created by interacting with a grass block using a hoe.
// Farmland takes into consideration its distance from a water source block to increase its efficiency and hold its hydration state
type Farmland struct {
	noNBT
	tilledGrass

	// Hydration is how much moisture a block has. The max and default hydration is 7, with its lowest
	// hydration being 0, which is essentially dirt with a crop on it.
	// This is calculated by checking if there is a water source block 4 blocks away in any direction from
	// the farmland that is either 1 block above or on the same level as the farmland block.
	// If there isn't, we then count down the hydration level by one until it eventually dries up and turns back into dirt.
	Hydration int
}

// NeighbourUpdateTick ...
func (f Farmland) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if solid := w.Block(pos.Side(world.FaceUp)).Model().FaceSolid(pos.Side(world.FaceUp), world.FaceDown, w); solid {
		w.SetBlock(pos, Dirt{})
	}
}

// RandomTick ...
func (f Farmland) RandomTick(pos world.BlockPos, w *world.World, _ *rand.Rand) {
	if !f.CanHydrate(pos, w) {
		if f.Hydration > 0 {
			f.Hydration--
			w.PlaceBlock(pos, f)
		} else {
			blockAbove := w.Block(pos.Side(world.FaceUp))
			if _, cropAbove := blockAbove.(Crop); !cropAbove {
				w.PlaceBlock(pos, Dirt{})
			}
		}
	} else {
		f.Hydration = 7
		w.PlaceBlock(pos, f)
	}
}

// CanHydrate checks for water within 4 blocks in each direction from the farmland.
func (f Farmland) CanHydrate(pos world.BlockPos, w *world.World) bool {
	for y := 0; y <= 1; y++ {
		for x := -4; x <= 4; x++ {
			for z := -4; z <= 4; z++ {
				if liquid, ok := w.Liquid(world.BlockPos{pos.X() + x, pos.Y() + y, pos.Z() + z}); ok {
					if _, ok := liquid.(Water); ok {
						return true
					}
				}
			}
		}
	}
	return false
}

// BreakInfo ...
func (f Farmland) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(Dirt{}, 1)),
	}
}

// EncodeBlock ...
func (f Farmland) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:farmland", map[string]interface{}{"moisturized_amount": int32(f.Hydration)}
}

func (f Farmland) Hash() uint64 {
	return hashFarmland | (uint64(f.Hydration) << 32)
}

// allFarmland returns all possible states that a block of farmland can be in.
func allFarmland() (b []world.Block) {
	for i := 0; i <= 7; i++ {
		b = append(b, Farmland{Hydration: i})
	}
	return
}
