package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// SwiftSneak is a non-renewable enchantment that can be applied to leggings and allows the player to walk more quickly
// while sneaking.
type SwiftSneak struct{}

// Name ...
func (SwiftSneak) Name() string {
	return "Swift Sneak"
}

// MaxLevel ...
func (SwiftSneak) MaxLevel() int {
	return 3
}

// MinCost ...
func (SwiftSneak) MinCost(level int) int {
	return level * 25
}

// MaxCost ...
func (s SwiftSneak) MaxCost(level int) int {
	return s.MinCost(level) + 50
}

// Rarity ...
func (SwiftSneak) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment ...
func (SwiftSneak) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// Treasure ...
func (SwiftSneak) Treasure() bool {
	return true
}

// CompatibleWithItem ...
func (SwiftSneak) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.LeggingsType)
	return ok && b.Leggings()
}
