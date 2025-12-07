package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Multishot is an enchantment for crossbows that shoots 3 arrows at once.
var Multishot multishot

type multishot struct{}

// Name ...
func (multishot) Name() string {
	return "multishot"
}

// MaxLevel ...
func (multishot) MaxLevel() int {
	return 1
}

// Cost ...
func (m multishot) Cost(level int) (int, int) {
	return 20, 50
}

// Rarity ...
func (multishot) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (multishot) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	// multishot is incompatible with Piercing
	return true
}

// CompatibleWithItem ...
func (multishot) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Crossbow)
	return ok
}

// MultipleProjectiles ...
func (multishot) MultipleProjectiles() bool {
	return true
}
