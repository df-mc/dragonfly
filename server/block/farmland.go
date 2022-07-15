package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
)

// Farmland is a block that crops are grown on. Farmland is created by interacting with a grass or dirt block using a
// hoe. Farmland can be hydrated by nearby water, with no hydration resulting in it turning into a dirt block.
type Farmland struct {
	tilledGrass

	// Hydration is how much moisture the farmland block has. Hydration starts at 0 & caps at 7. During a random tick
	// update, if there is water within 4 blocks from the farmland block, hydration is set to 7. Otherwise, it
	// decrements until it turns into dirt.
	Hydration int
}

// SoilFor ...
func (f Farmland) SoilFor(block world.Block) bool {
	switch block.(type) {
	case TallGrass, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts:
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (f Farmland) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if solid := w.Block(pos.Side(cube.FaceUp)).Model().FaceSolid(pos.Side(cube.FaceUp), cube.FaceDown, w); solid {
		w.SetBlock(pos, Dirt{}, nil)
	}
}

// RandomTick ...
func (f Farmland) RandomTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if !f.hydrated(pos, w) {
		if f.Hydration > 0 {
			f.Hydration--
			w.SetBlock(pos, f, nil)
		} else {
			blockAbove := w.Block(pos.Side(cube.FaceUp))
			if _, cropAbove := blockAbove.(Crop); !cropAbove {
				w.SetBlock(pos, Dirt{}, nil)
			}
		}
	} else {
		f.Hydration = 7
		w.SetBlock(pos, f, nil)
	}
}

// hydrated checks for water within 4 blocks in each direction from the farmland.
func (f Farmland) hydrated(pos cube.Pos, w *world.World) bool {
	posX, posY, posZ := pos.X(), pos.Y(), pos.Z()
	for y := 0; y <= 1; y++ {
		for x := -4; x <= 4; x++ {
			for z := -4; z <= 4; z++ {
				if liquid, ok := w.Liquid(cube.Pos{posX + x, posY + y, posZ + z}); ok {
					if _, ok := liquid.(Water); ok {
						return true
					}
				}
			}
		}
	}
	return false
}

// EntityLand ...
func (f Farmland) EntityLand(pos cube.Pos, w *world.World, e world.Entity, distance *float64) {
	if living, ok := e.(entity.Living); ok {
		if fall, ok := living.(fallDistanceEntity); ok && rand.Float64() < fall.FallDistance()-0.5 {
			w.SetBlock(pos, Dirt{}, nil)
		}
	}
}

// fallDistanceEntity is an entity that has a fall distance.
type fallDistanceEntity interface {
	// ResetFallDistance resets the entities fall distance.
	ResetFallDistance()
	// FallDistance returns the entities fall distance.
	FallDistance() float64
}

// BreakInfo ...
func (f Farmland) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, oneOf(Dirt{}))
}

// EncodeBlock ...
func (f Farmland) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:farmland", map[string]any{"moisturized_amount": int32(f.Hydration)}
}

// EncodeItem ...
func (f Farmland) EncodeItem() (name string, meta int16) {
	return "minecraft:farmland", 0
}

// allFarmland returns all possible states that a block of farmland can be in.
func allFarmland() (b []world.Block) {
	for i := 0; i <= 7; i++ {
		b = append(b, Farmland{Hydration: i})
	}
	return
}
