package block

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"math"
)

// Crop is an interface for all crops that are grown on farmland. A crop has a random chance to grow during random ticks.
type Crop interface {
	// GrowthStage returns the crop's current stage of growth.
	GrowthStage() int
}

// crop is a base for crop plants.
type crop struct {
	noNBT
	transparent
	empty

	// Growth is the current stage of growth.
	Growth int
}

// GrowthStage returns the current stage of growth.
func (c crop) GrowthStage() int {
	return c.Growth
}

// CalculateGrowthChance calculates the chance the crop will grow during a random tick.
func (c crop) CalculateGrowthChance(crop Crop, pos world.BlockPos, w *world.World) float64 {
	points := 0

	under := pos.Side(world.FaceDown)

	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			block := w.Block(under.Add(world.BlockPos{x, 0, z}))
			if farmland, ok := block.(Farmland); ok {
				farmlandPoints := 0
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

	chance := 1 / (math.Floor(float64(25/points)) + 1)
	return chance
}
