package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// SwiftSneak is a non-renewable enchantment that can be applied to leggings
// and allows the player to walk more quickly while sneaking.
var SwiftSneak swiftSneak

type swiftSneak struct{}

func (swiftSneak) Name() string {
	return "Swift Sneak"
}

func (swiftSneak) MaxLevel() int {
	return 3
}

func (swiftSneak) Cost(level int) (int, int) {
	minCost := level * 25
	return minCost, minCost + 50
}

func (swiftSneak) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityVeryRare
}

func (swiftSneak) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

func (swiftSneak) Treasure() bool {
	return true
}

func (swiftSneak) CompatibleWithItem(i world.Item) bool {
	b, ok := i.(item.LeggingsType)
	return ok && b.Leggings()
}
