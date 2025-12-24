package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Multishot is an enchantment for crossbows that allow them to shoot three arrows or firework rockets at the cost of one.
var Multishot multishot

type multishot struct{}

// Name ...
func (multishot) Name() string {
	return "Multishot"
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
	// TODO: Multishot is incompatible with Piercing
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
