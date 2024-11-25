package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Efficiency is an enchantment that increases mining speed.
var Efficiency efficiency

type efficiency struct{}

// Name ...
func (efficiency) Name() string {
	return "Efficiency"
}

// MaxLevel ...
func (efficiency) MaxLevel() int {
	return 5
}

// Cost ...
func (efficiency) Cost(level int) (int, int) {
	minCost := 1 + 10*(level-1)
	return minCost, minCost + 50
}

// Rarity ...
func (efficiency) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// Addend returns the mining speed addend from efficiency.
func (efficiency) Addend(level int) float64 {
	return float64(level*level + 1)
}

// CompatibleWithEnchantment ...
func (efficiency) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (efficiency) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() != item.TypeSword && t.ToolType() != item.TypeNone)
}
