package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Efficiency is an enchantment that increases mining speed.
type Efficiency struct{}

// Addend returns the mining speed addend from efficiency.
func (e Efficiency) Addend(level int) float64 {
	return float64(level*level + 1)
}

// Name ...
func (e Efficiency) Name() string {
	return "Efficiency"
}

// MaxLevel ...
func (e Efficiency) MaxLevel() int {
	return 5
}

// Rarity ...
func (e Efficiency) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// CompatibleWithOther ...
func (e Efficiency) CompatibleWithOther(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (e Efficiency) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() == item.TypeAxe || t.ToolType() == item.TypePickaxe || t.ToolType() == item.TypeShovel)
}
