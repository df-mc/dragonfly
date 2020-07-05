package block

import (
	"github.com/df-mc/dragonfly/dragonfly/internal/item_internal"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"math/rand"
)

type Farmland struct {
	// Hydration is how much moisture a block has. This is calculated by checking if there is a water source block
	// 4 blocks away in any direction from the farmland that is either 1 block above or on the same level as the farmland block.
	// If there isn't, we then count down the hydration level by one until it eventually dries up and turns back into dirt.
	Hydration uint8
}

// NeighbourUpdateTick ...
func (f Farmland) NeighbourUpdateTick(pos, block world.BlockPos, w *world.World) {
	if _, isAir := w.Block(pos.Side(world.FaceUp)).(Air); !isAir {
		if _, isCrop := w.Block(pos.Side(world.FaceUp)).(Crop); !isCrop {
			w.SetBlock(pos, Dirt{})
		}
	}
}

// RandomTick ...
func (f Farmland) RandomTick(pos world.BlockPos, w *world.World, r *rand.Rand) {
	// Calculate the Hydration of the farmland block.
	f.CalculateHydration(pos, w)
	// Check if there is a crop on the farmland block.
	if crop, isCrop := w.Block(pos.Add(world.BlockPos{0, 1})).(Crop); isCrop {
		// Check if the crop requires Hydration to grow and if it meets those requirements.
		if crop.RequiresHydration() && f.Hydration == 0 {
			return
		}

		// Check if the crop can grow due to lighting.
		if w.Light(pos.Add(world.BlockPos{0, 1})) >= crop.LightLevelRequired() {
			crop.Grow(pos.Add(world.BlockPos{0, 1}), w, r, f.Hydration)
		}
	} else if f.Hydration <= 0 {
		// If no crop exists and the Hydration level is 0, turn the block into dirt.
		w.SetBlock(pos, item_internal.Dirt)
	}

	// If the farmland is dehydrated it should be replaced with it's corresponding farmland block.
	if f.Hydration < 7 && f.Hydration > 0 {
		// Decrease the hydration one by one down to zero.
		w.SetBlock(pos, Farmland{f.Hydration - 1})
	}
}

// CalculateHydration determines the Hydration or moisture of a block by scanning a 4 block box
// around the Farmland block in each direction for water source blocks.
// This also takes into account water source blocks one block higher than the farmland.
func (f Farmland) CalculateHydration(pos world.BlockPos, w *world.World) {
	// Start on the original Y level of the farmland block and make 9x9 with the center being the farmland block.
	// If any water source blocks are found, return max hydration.
	for yLevel := 0; yLevel <= 1; yLevel++ {
		for xLevel := -4; xLevel <= 4; xLevel++ {
			for zLevel := -4; zLevel <= 4; zLevel++ {
				if water, isWater := w.Block(world.BlockPos{pos.X() + xLevel, pos.Y() + yLevel, pos.Z() + zLevel}).(Water); isWater && water.Depth == 8 {
					// If the blocks Hydration wasn't 7 before, then replace the block.
					if f.Hydration < 7 {
						w.SetBlock(pos, Farmland{7})
					}
					return
				}
			}
		}
	}
	// Checks if the farmland block was previously hydrated.
	if f.Hydration == 7 {
		// No water blocks are found, meaning the block is now a dehydrated farmland block.
		w.SetBlock(pos, Farmland{6})
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

// allFarmland returns all possible states that a block of farmland can be in.
func allFarmland() (b []world.Block) {
	for i := 7; i >= 0; i-- {
		b = append(b, Farmland{Hydration: uint8(i)})
	}
	return
}
