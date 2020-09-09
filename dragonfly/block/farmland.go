package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"math/rand"
)

// Farmland is a block that crops are grown on. Farmland is created by interacting with a grass or dirt block using a
// hoe. Farmland can be hydrated by nearby water, with no hydration resulting in it turning into a dirt block.
type Farmland struct {
	noNBT
	tilledGrass

	// Hydration is how much moisture the farmland block has. Hydration starts at 0 & caps at 7. During a random tick
	// update, if there is water within 4 blocks from the farmland block, hydration is set to 7. Otherwise, it
	// decrements until it turns into dirt.
	Hydration int
}

//TODO: Add crop trampling

// NeighbourUpdateTick ...
func (f Farmland) NeighbourUpdateTick(pos, _ world.BlockPos, w *world.World) {
	if solid := w.Block(pos.Side(world.FaceUp)).Model().FaceSolid(pos.Side(world.FaceUp), world.FaceDown, w); solid {
		w.SetBlock(pos, Dirt{})
	}
}

// RandomTick ...
func (f Farmland) RandomTick(pos world.BlockPos, w *world.World, _ *rand.Rand) {
	if !f.hydrated(pos, w) {
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

// hydrated checks for water within 4 blocks in each direction from the farmland.
func (f Farmland) hydrated(pos world.BlockPos, w *world.World) bool {
	posX := pos.X()
	posY := pos.Y()
	posZ := pos.Z()
	for y := 0; y <= 1; y++ {
		for x := -4; x <= 4; x++ {
			for z := -4; z <= 4; z++ {
				if liquid, ok := w.Liquid(world.BlockPos{posX + x, posY + y, posZ + z}); ok {
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

// Hash ...
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
