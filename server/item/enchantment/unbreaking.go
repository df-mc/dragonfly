package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand/v2"
)

// Unbreaking is an enchantment that gives a chance for an item to avoid
// durability reduction when it is used, effectively increasing the item's
// durability.
var Unbreaking unbreaking

type unbreaking struct{}

// Name ...
func (unbreaking) Name() string {
	return "Unbreaking"
}

// MaxLevel ...
func (unbreaking) MaxLevel() int {
	return 3
}

// Cost ...
func (unbreaking) Cost(level int) (int, int) {
	minCost := 5 + 8*(level-1)
	return minCost, minCost + 50
}

// Rarity ...
func (unbreaking) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// CompatibleWithEnchantment ...
func (unbreaking) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (unbreaking) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Durable)
	return ok
}

// Reduce returns the amount of damage that should be reduced with unbreaking.
func (unbreaking) Reduce(it world.Item, level, amount int) int {
	after := amount
	_, ok := it.(item.Armour)
	for i := 0; i < amount; i++ {
		if (!ok || rand.Float64() >= 0.6) && rand.IntN(level+1) > 0 {
			after--
		}
	}
	return after
}
