package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Mending is an enchantment that repairs the item when experience orbs are
// collected.
var Mending mending

type mending struct{}

// Name ...
func (mending) Name() string {
	return "Mending"
}

// MaxLevel ...
func (mending) MaxLevel() int {
	return 1
}

// Cost ...
func (mending) Cost(level int) (int, int) {
	minCost := level * 25
	return minCost, minCost + 50
}

// Rarity ...
func (mending) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// Treasure ...
func (mending) Treasure() bool {
	return true
}

// CompatibleWithEnchantment ...
func (mending) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Infinity
}

// CompatibleWithItem ...
func (mending) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Durable)
	return ok
}
