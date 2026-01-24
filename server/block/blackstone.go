package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Blackstone is a naturally generating block in the nether that can be used to craft stone tools, brewing stands and
// furnaces. Gilded blackstone also has a 10% chance to drop 2-6 golden nuggets.
type Blackstone struct {
	solid
	bassDrum

	// Type is the type of blackstone of the block.
	Type BlackstoneType
}

// BreakInfo ...
func (b Blackstone) BreakInfo() BreakInfo {
	drops := oneOf(b)
	hardness := 1.5

	switch b.Type {
	case GildedBlackstone():
		drops = func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
			if hasSilkTouch(enchantments) {
				return []item.Stack{item.NewStack(b, 1)}
			}
			nuggetChances := []float64{0.1, 1.0 / 7.0, 0.25, 1.0}
			if rand.Float64() < nuggetChances[min(fortuneLevel(enchantments), 3)] {
				return []item.Stack{item.NewStack(item.GoldNugget{}, rand.IntN(4)+2)}
			}
			return []item.Stack{item.NewStack(b, 1)}
		}
	case PolishedBlackstone():
		hardness = 2
	}

	return newBreakInfo(hardness, pickaxeHarvestable, pickaxeEffective, drops).withBlastResistance(30)
}

// EncodeItem ...
func (b Blackstone) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Type.String(), 0
}

// EncodeBlock ...
func (b Blackstone) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + b.Type.String(), nil
}

// allBlackstone returns a list of all blackstone block variants.
func allBlackstone() (s []world.Block) {
	for _, t := range BlackstoneTypes() {
		s = append(s, Blackstone{Type: t})
	}
	return
}
