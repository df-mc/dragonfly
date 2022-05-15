package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
)

// Grass blocks generate abundantly across the surface of the world.
type Grass struct {
	solid
}

// plantSelection are the plants that are picked from when a bone meal is attempted.
// TODO: Base plant selection on current biome.
var plantSelection = []world.Block{
	Flower{Type: OxeyeDaisy()},
	Flower{Type: PinkTulip()},
	Flower{Type: Cornflower()},
	Flower{Type: WhiteTulip()},
	Flower{Type: RedTulip()},
	Flower{Type: OrangeTulip()},
	Flower{Type: Dandelion()},
	Flower{Type: Poppy()},
}

// init adds extra variants of TallGrass to the plant selection.
func init() {
	for i := 0; i < 8; i++ {
		plantSelection = append(plantSelection, TallGrass{Type: Fern()})
	}
	for i := 0; i < 12; i++ {
		plantSelection = append(plantSelection, TallGrass{Type: NormalGrass()})
	}
}

// SoilFor ...
func (g Grass) SoilFor(block world.Block) bool {
	switch block.(type) {
	case TallGrass, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts:
		return true
	}
	return false
}

// RandomTick handles the ticking of grass, which may or may not result in the spreading of grass onto dirt.
func (g Grass) RandomTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	aboveLight := w.Light(pos.Add(cube.Pos{0, 1}))
	if aboveLight < 4 {
		// The light above the block is too low: The grass turns to dirt.
		w.SetBlock(pos, Dirt{}, nil)
		return
	}
	if aboveLight < 9 {
		// Don't attempt to spread if the light level is lower than 9.
		return
	}

	// Generate a single uint32 as we only need 28 bits (7 bits each iteration).
	n := r.Uint32()

	// Four attempts to spread to another block.
	for i := 0; i < 4; i++ {
		x, y, z := int(n)%3, int(n>>2)%5, int(n>>5)%3
		n >>= 7

		spreadPos := pos.Add(cube.Pos{x - 1, y - 3, z - 1})
		b := w.Block(spreadPos)
		if dirt, ok := b.(Dirt); !ok || dirt.Coarse {
			continue
		}
		// Don't spread grass to locations where dirt is exposed to hardly any light.
		if w.Light(spreadPos.Add(cube.Pos{0, 1})) < 4 {
			continue
		}
		w.SetBlock(spreadPos, g, nil)
	}
}

// BoneMeal ...
func (g Grass) BoneMeal(pos cube.Pos, w *world.World) bool {
	for c := 0; c < 14; c++ {
		x := randWithinRange(pos.X()-3, pos.X()+3)
		z := randWithinRange(pos.Z()-3, pos.Z()+3)
		if (w.Block(cube.Pos{x, pos.Y() + 1, z}) == Air{}) && (w.Block(cube.Pos{x, pos.Y(), z}) == Grass{}) {
			w.SetBlock(cube.Pos{x, pos.Y() + 1, z}, plantSelection[randWithinRange(0, len(plantSelection)-1)], nil)
		}
	}

	return false
}

// BreakInfo ...
func (g Grass) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, silkTouchOneOf(Dirt{}, g))
}

// EncodeItem ...
func (Grass) EncodeItem() (name string, meta int16) {
	return "minecraft:grass", 0
}

// EncodeBlock ...
func (Grass) EncodeBlock() (string, map[string]any) {
	return "minecraft:grass", nil
}

// Till ...
func (g Grass) Till() (world.Block, bool) {
	return Farmland{}, true
}

// Shovel ...
func (g Grass) Shovel() (world.Block, bool) {
	return DirtPath{}, true
}

// randWithinRange returns a random integer within a range.
func randWithinRange(min, max int) int {
	return rand.Intn(max-min) + min
}
