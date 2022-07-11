package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
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
func (Mending) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	//_, infinity := s.Enchantment(Infinity{}) todo: infinity
	// return !infinty
	return true
}

// CompatibleWithItem ...
func (Mending) CompatibleWithItem(s item.Stack) bool {
	_, ok := s.Item().(item.Durable)
	return ok
}
