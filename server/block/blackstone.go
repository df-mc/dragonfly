package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
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
	if b.Type == GildedBlackstone() {
		drops = func(item.Tool, []item.Enchantment) []item.Stack {
			if rand.Float64() < 0.1 {
				return []item.Stack{item.NewStack(item.GoldNugget{}, rand.Intn(4)+2)}
			}
			return []item.Stack{item.NewStack(b, 1)}
		}
	}
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, drops).withBlastResistance(30)
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
