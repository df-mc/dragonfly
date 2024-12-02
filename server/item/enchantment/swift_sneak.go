package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// SwiftSneak is a non-renewable enchantment that can be applied to leggings
// and allows the player to walk more quickly while sneaking.
var SwiftSneak swiftSneak

type swiftSneak struct{}

// Name ...
func (swiftSneak) Name() string {
	return "Swift Sneak"
}

// MaxLevel ...
func (swiftSneak) MaxLevel() int {
	return 3
}

// Cost ...
func (swiftSneak) Cost(level int) (int, int) {
	minCost := level * 25
	return minCost, minCost + 50
}

// Rarity ...
func (swiftSneak) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

// CompatibleWithEnchantment ...
func (swiftSneak) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// Treasure ...
func (swiftSneak) Treasure() bool {
	return true
}

// CompatibleWithItem ...
func (swiftSneak) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.LeggingsType)
	return ok && b.Leggings()
}
