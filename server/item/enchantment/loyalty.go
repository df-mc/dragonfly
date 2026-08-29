package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Loyalty is a trident enchantment that causes a thrown trident to return to its owner
// after hitting a block or an entity.
var Loyalty loyalty

type loyalty struct{}

// Name ...
func (loyalty) Name() string {
	return "Loyalty"
}

// MaxLevel ...
func (loyalty) MaxLevel() int {
	return 3
}

// Cost ...
func (loyalty) Cost(level int) (int, int) {
	return 5 + level*7, 50
}

// Rarity ...
func (loyalty) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// CompatibleWithEnchantment ...
func (loyalty) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != Riptide
}

// CompatibleWithItem ...
func (loyalty) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Trident)
	return ok
}
