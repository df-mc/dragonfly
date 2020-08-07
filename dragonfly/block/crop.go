package block

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Crop is an interface for all crops that are grown on farmland. A crop has a random chance to grow during random ticks.
type Crop interface {
	// GrowthStage returns the crop's current stage of growth. The max value is 7.
	GrowthStage() int
}

// crop is a base for crop plants.
type crop struct {
	noNBT
	transparent
	empty

	// Growth is the current stage of growth. The max value is 7.
	Growth int
}

// NeighbourUpdateTick ...
func (c crop) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if _, ok := w.Block(pos.Side(world.FaceDown)).(Farmland); !ok {
		w.BreakBlock(pos)
	}
}

// HasLiquidDrops ...
func (c crop) HasLiquidDrops() bool {
	return true
}

// GrowthStage returns the current stage of growth.
func (c crop) GrowthStage() int {
	return c.Growth
}

// CalculateGrowthChance calculates the chance the crop will grow during a random tick.
func (c crop) CalculateGrowthChance(pos world.BlockPos, w *world.World) float64 {
	points := 0.0

	block := w.Block(pos)
	under := pos.Side(world.FaceDown)

	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			block := w.Block(under.Add(world.BlockPos{x, 0, z}))
			if farmland, ok := block.(Farmland); ok {
				farmlandPoints := 0.0
				if farmland.Hydration > 0 {
					farmlandPoints = 4
				} else {
					farmlandPoints = 2
				}
				if x != 0 || z != 0 {
					farmlandPoints = (farmlandPoints - 1) / 4
				}
				points += farmlandPoints
			}
		}
	}

	north := pos.Side(world.FaceNorth)
	south := pos.Side(world.FaceSouth)

	northSouth := sameCrop(block, w.Block(north)) || sameCrop(block, w.Block(south))
	westEast := sameCrop(block, w.Block(pos.Side(world.FaceWest))) || sameCrop(block, w.Block(pos.Side(world.FaceEast)))
	if northSouth && westEast {
		points /= 2
	} else {
		diagonal := sameCrop(block, w.Block(north.Side(world.FaceWest))) ||
			sameCrop(block, w.Block(north.Side(world.FaceEast))) ||
			sameCrop(block, w.Block(south.Side(world.FaceWest))) ||
			sameCrop(block, w.Block(south.Side(world.FaceEast)))
		if diagonal {
			points /= 2
		}
	}

	chance := 1 / (25/points + 1)
	return chance
}

// sameCrop checks if both blocks are crops and that they are the same type.
func sameCrop(blockA, blockB world.Block) bool {
	if _, ok := blockA.(Crop); ok {
		if _, ok := blockB.(Crop); ok {
			nameA, _ := blockA.EncodeBlock()
			nameB, _ := blockB.EncodeBlock()
			return nameA == nameB
		}
	}
	return false
}

// min returns the smaller of the two integers passed.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
