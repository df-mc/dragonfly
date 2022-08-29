package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Vanishing is an enchantment that causes the item to disappear on death.
type Vanishing struct{}

// Name ...
func (Vanishing) Name() string {
	return "Vanishing"
}

// MaxLevel ...
func (Vanishing) MaxLevel() int {
	return 1
}

// Cost ...
func (Vanishing) Cost(int) (int, int) {
	return 15, 65
}

// Rarity ...
func (Vanishing) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment...
func (Vanishing) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem...
func (Vanishing) CompatibleWithItem(i world.Item) bool {
	return true
}

// Treasure ...
func (Vanishing) Treasure() bool {
	return true
}
