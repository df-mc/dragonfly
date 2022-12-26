package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// FrostWalker is an enchantment to boots that creates frosted ice blocks when walking
// over water, and causes the wearer to be immune to damage from blocks such as magma
// blocks and campfires when stepped on.
type FrostWalker struct{}

// Name ...
func (FrostWalker) Name() string {
	return "Frost Walker"
}

// MaxLevel ...
func (FrostWalker) MaxLevel() int {
	return 2
}

// Cost ...
func (FrostWalker) Cost(level int) (int, int) {
	min := level * 10
	return min, min + 15
}

// Rarity ...
func (FrostWalker) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment ...
func (FrostWalker) CompatibleWithEnchantment(et item.EnchantmentType) bool {
	_, isDepthStrider := et.(DepthStrider)
	return !isDepthStrider
}

// CompatibleWithItem ...
func (FrostWalker) CompatibleWithItem(i world.Item) bool {
	_, isBoots := i.(item.Boots)

	return isBoots
}

// Treasure ...
func (FrostWalker) Treasure() bool {
	return true
}
