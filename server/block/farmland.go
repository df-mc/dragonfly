package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
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
		w.SetBlock(pos, Dirt{})
	}
}

// RandomTick ...
func (f Farmland) RandomTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if !f.hydrated(pos, w) {
		if f.Hydration > 0 {
			f.Hydration--
			w.PlaceBlock(pos, f)
		} else {
			blockAbove := w.Block(pos.Side(cube.FaceUp))
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

// BreakInfo ...
func (f Farmland) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, oneOf(Dirt{}))
}

// EncodeBlock ...
func (f Farmland) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:farmland", map[string]interface{}{"moisturized_amount": int32(f.Hydration)}
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

/*
CloudBurst uses 0.75, but using the entity collide method I can't reliably reach the minimum of 0.75
I am able however to always reach 0.37, so I went with that instead. Should be good enough.
*/
const minimumFallDistance = 0.37

// EntityCollide ...
func (f Farmland) EntityCollide(pos cube.Pos, e world.Entity) {
	if fallEntity, ok := e.(FallDistanceEntity); ok {
		fallDistance := fallEntity.FallDistance()
		if fallDistance > minimumFallDistance {
			e.World().PlaceBlock(pos, Dirt{})
		}
	}
}
