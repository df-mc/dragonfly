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
	return 25, 50
}

// Rarity ...
func (Vanishing) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment ...
func (Vanishing) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (Vanishing) CompatibleWithItem(i world.Item) bool {
	_, dur := i.(item.Durable)
	_, com := i.(item.Compass)
	_, arm := i.(item.Armour)
	// TODO: Recovery Compass
	// TODO: Carrot on a Stick
	// TODO: Warped Fungus on a Stick
	return arm || com || dur
}

// Treasure ...
func (Vanishing) Treasure() bool {
	return true
}

// Treasure ...
func (Vanishing) Curse() bool {
	return true
}
