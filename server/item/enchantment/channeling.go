package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Channeling is a trident enchantment that produces lightning when thrown at a mob or lightning rod while a
// thunderstorm is occurring.
var Channeling channeling

type channeling struct{}

// Name ...
func (channeling) Name() string {
	return "Channeling"
}

// MaxLevel ...
func (channeling) MaxLevel() int {
	return 1
}

// Cost ...
func (channeling) Cost(int) (int, int) {
	return 25, 50
}

// Rarity ...
func (channeling) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment ...
func (channeling) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Riptide
}

// CompatibleWithItem ...
func (channeling) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Trident)
	return ok
}
