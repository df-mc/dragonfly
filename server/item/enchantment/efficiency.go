package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Efficiency is an enchantment that increases mining speed.
var Efficiency efficiency

type efficiency struct{}

func (efficiency) Name() string {
	return "Efficiency"
}

func (efficiency) MaxLevel() int {
	return 5
}

func (efficiency) Cost(level int) (int, int) {
	minCost := 1 + 10*(level-1)
	return minCost, minCost + 50
}

func (efficiency) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// Addend returns the mining speed addend from efficiency.
func (efficiency) Addend(level int) float64 {
	return float64(level*level + 1)
}

func (efficiency) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (efficiency) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() != item.TypeSword && t.ToolType() != item.TypeNone)
}
