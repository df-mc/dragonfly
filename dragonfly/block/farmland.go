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
func (f Farmland) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	// Calculate the Hydration of the farmland block.
	f.Hydrate(pos, w)

	// Check if there is a crop on the farmland block.
	if _, isCrop := w.Block(pos.Side(world.FaceUp)).(Crop); !isCrop && f.Hydration == 0 {
		// If no crop exists and the Hydration level is 0, turn the block into dirt.
		w.SetBlock(pos, Dirt{})
	}
}

// Hydrate determines the Hydration or moisture of a block by scanning a 4 block box
// around the Farmland block in each direction for water source blocks.
// This also takes into account water source blocks one block higher than the farmland.
func (f Farmland) Hydrate(pos world.BlockPos, w *world.World) {
	// Start on the original Y level of the farmland block and make 9x9 with the center being the farmland block.
	// If any water source blocks are found, return max hydration.
	for y := 0; y <= 1; y++ {
		for x := -4; x <= 4; x++ {
			for z := -4; z <= 4; z++ {
				if liquid, ok := w.Liquid(world.BlockPos{pos.X() + x, pos.Y() + y, pos.Z() + z}); ok {
					if _, ok := liquid.(Water); ok {
						// If the blocks Hydration wasn't 7 before, then replace the block.
						if f.Hydration < 7 {
							w.SetBlock(pos, Farmland{Hydration: 7})
						}
						return
					}
				}
			}
		}
	}
	// Checks if the farmland block was previously hydrated.
	if f.Hydration > 0 {
		// No water blocks are found, meaning the block is now a dehydrated farmland block.
		w.SetBlock(pos, Farmland{Hydration: f.Hydration - 1})
	}
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
