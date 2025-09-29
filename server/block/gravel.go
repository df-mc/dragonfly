package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand/v2"
)

// Gravel is a block affected by gravity. It has a 10% chance of dropping flint instead of itself on break.
type Gravel struct {
	gravityAffected
	solid
	snare
}

func (g Gravel) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	g.fall(g, pos, tx)
}

func (g Gravel) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if !hasSilkTouch(enchantments) && rand.Float64() < 0.1 {
			return []item.Stack{item.NewStack(item.Flint{}, 1)}
		}
		return []item.Stack{item.NewStack(g, 1)}
	})
}

func (Gravel) EncodeItem() (name string, meta int16) {
	return "minecraft:gravel", 0
}

func (Gravel) EncodeBlock() (string, map[string]any) {
	return "minecraft:gravel", nil
}
