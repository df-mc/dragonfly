package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Mending is an enchantment that repairs the item when experience orbs are
// collected.
var Mending mending

type mending struct{}

func (mending) Name() string {
	return "Mending"
}

func (mending) MaxLevel() int {
	return 1
}

func (mending) Cost(level int) (int, int) {
	minCost := level * 25
	return minCost, minCost + 50
}

func (mending) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

func (mending) Treasure() bool {
	return true
}

func (mending) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Infinity
}

func (mending) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Durable)
	return ok
}
