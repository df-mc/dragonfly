package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Gravel is a block affected by gravity. It has a 10% chance of dropping flint instead of itself on break.
type Gravel struct {
	gravityAffected
	solid
	snare
}

// NeighbourUpdateTick ...
func (g Gravel) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	g.fall(g, pos, tx)
}

// BreakInfo ...
func (g Gravel) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(g, 1)}
		}
		flintChances := []float64{0.1, 1.0 / 7.0, 0.25, 1.0}
		flintChance := flintChances[min(fortuneLevel(enchantments), 3)]
		if rand.Float64() < flintChance {
			return []item.Stack{item.NewStack(item.Flint{}, 1)}
		}
		return []item.Stack{item.NewStack(g, 1)}
	})
}

// EncodeItem ...
func (Gravel) EncodeItem() (name string, meta int16) {
	return "minecraft:gravel", 0
}

// EncodeBlock ...
func (Gravel) EncodeBlock() (string, map[string]any) {
	return "minecraft:gravel", nil
}
