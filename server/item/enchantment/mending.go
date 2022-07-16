package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Mending is an enchantment that repairs the item when experience orbs are collected.
type Mending struct{}

// Name ...
func (Mending) Name() string {
	return "Mending"
}

// MaxLevel ...
func (Mending) MaxLevel() int {
	return 1
}

// Rarity ...
func (Mending) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (Mending) CompatibleWithEnchantment(item.EnchantmentType) bool {
	//_, infinity := t.(Infinity) TODO: infinity
	// return !infinty
	return true
}

// CompatibleWithItem ...
func (Mending) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Durable)
	return ok
}
