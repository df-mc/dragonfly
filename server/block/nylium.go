package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Nylium is a variant of netherrack that covers the floor of crimson and warped forests, on which nether
// vegetation is able to grow.
type Nylium struct {
	solid
	bassDrum

	// Warped is the turquoise variant found in warped forests.
	Warped bool
}

// SoilFor ...
func (n Nylium) SoilFor(block world.Block) bool {
	_, ok := block.(NetherSprouts)
	return ok
}

// RandomTick ...
func (n Nylium) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	above := pos.Side(cube.FaceUp)
	if diffuser, ok := tx.Block(above).(LightDiffuser); !ok || diffuser.LightDiffusionLevel() == 15 {
		tx.SetBlock(pos, Netherrack{}, nil)
	}
}

// BreakInfo ...
func (n Nylium) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(Netherrack{}, n))
}

// EncodeItem ...
func (n Nylium) EncodeItem() (name string, meta int16) {
	if n.Warped {
		return "minecraft:warped_nylium", 0
	}
	return "minecraft:crimson_nylium", 0
}

// EncodeBlock ...
func (n Nylium) EncodeBlock() (string, map[string]any) {
	if n.Warped {
		return "minecraft:warped_nylium", nil
	}
	return "minecraft:crimson_nylium", nil
}
