package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Infinity is an enchantment to bows that prevents regular arrows from being consumed when shot.
type Infinity struct{}

// Name ...
func (Infinity) Name() string {
	return "Infinity"
}

// MaxLevel ...
func (Infinity) MaxLevel() int {
	return 1
}

// Rarity ...
func (Infinity) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// ConsumesArrows always returns false.
func (Infinity) ConsumesArrows() bool {
	return false
}

// CompatibleWithEnchantment ...
func (Infinity) CompatibleWithEnchantment(item.EnchantmentType) bool {
	// TODO: Account for mending.
	return true
}

// CompatibleWithItem ...
func (Infinity) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Bow)
	return ok
}
