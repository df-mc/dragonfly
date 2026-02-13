package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Crop is an interface for all crops that are grown on farmland. A crop has a random chance to grow during random ticks.
type Crop interface {
	// GrowthStage returns the crop's current stage of growth. The max value is 7.
	GrowthStage() int
	// SameCrop checks if two crops are of the same type.
	SameCrop(c Crop) bool
}

// crop is a base for crop plants.
type crop struct {
	transparent
	empty

	// Growth is the current stage of growth. The max value is 7.
	Growth int
}

// NeighbourUpdateTick ...
func (c crop) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Farmland); !ok {
		breakBlock(tx.Block(pos), pos, tx)
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
func (c crop) CalculateGrowthChance(pos cube.Pos, tx *world.Tx) float64 {
	points := 0.0

	block := tx.Block(pos)
	under := pos.Side(cube.FaceDown)

	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			block := tx.Block(under.Add(cube.Pos{x, 0, z}))
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

	north := pos.Side(cube.FaceNorth)
	south := pos.Side(cube.FaceSouth)

	northSouth := sameCrop(block, tx.Block(north)) || sameCrop(block, tx.Block(south))
	westEast := sameCrop(block, tx.Block(pos.Side(cube.FaceWest))) || sameCrop(block, tx.Block(pos.Side(cube.FaceEast)))
	if northSouth && westEast {
		points /= 2
	} else {
		diagonal := sameCrop(block, tx.Block(north.Side(cube.FaceWest))) ||
			sameCrop(block, tx.Block(north.Side(cube.FaceEast))) ||
			sameCrop(block, tx.Block(south.Side(cube.FaceWest))) ||
			sameCrop(block, tx.Block(south.Side(cube.FaceEast)))
		if diagonal {
			points /= 2
		}
	}

	chance := 1 / (25/points + 1)
	return chance
}

// sameCrop checks if both blocks are crops and that they are the same type.
func sameCrop(blockA, blockB world.Block) bool {
	if a, ok := blockA.(Crop); ok {
		if b, ok := blockB.(Crop); ok {
			return a.SameCrop(b)
		}
	}
	return false
}
