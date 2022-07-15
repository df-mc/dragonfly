package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Efficiency is an enchantment that increases mining speed.
type Efficiency struct{}

// Name ...
func (Efficiency) Name() string {
	return "Efficiency"
}

// MaxLevel ...
func (Efficiency) MaxLevel() int {
	return 5
}

// Rarity ...
func (Efficiency) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// Addend returns the mining speed addend from efficiency.
func (Efficiency) Addend(level int) float64 {
	return float64(level*level + 1)
}

// CompatibleWithEnchantment ...
func (Efficiency) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (Efficiency) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() != item.TypeSword && t.ToolType() != item.TypeNone)
}
