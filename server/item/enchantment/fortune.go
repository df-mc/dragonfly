package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Fortune is an enchantment that gives a chance to receive more item drops from certain blocks.
var Fortune fortune

type fortune struct{}

// Name ...
func (fortune) Name() string {
	return "Fortune"
}

// MaxLevel ...
func (fortune) MaxLevel() int {
	return 3
}

// Cost ...
func (fortune) Cost(level int) (int, int) {
	minCost := 15 + (level-1)*9
	return minCost, minCost + 50 + level
}

// Rarity ...
func (fortune) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (fortune) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != SilkTouch
}

// CompatibleWithItem ...
func (fortune) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() == item.TypePickaxe || t.ToolType() == item.TypeShovel || t.ToolType() == item.TypeAxe || t.ToolType() == item.TypeHoe)
}
