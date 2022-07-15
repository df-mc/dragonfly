package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
)

// Unbreaking is an enchantment that gives a chance for an item to avoid durability reduction when it
// is used, effectively increasing the item's durability.
type Unbreaking struct{}

// Name ...
func (Unbreaking) Name() string {
	return "Unbreaking"
}

// MaxLevel ...
func (Unbreaking) MaxLevel() int {
	return 3
}

// Rarity ...
func (Unbreaking) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Reduce returns the amount of damage that should be reduced with unbreaking.
func (Unbreaking) Reduce(it world.Item, level, amount int) int {
	after := amount
	_, ok := it.(item.Armour)
	for i := 0; i < amount; i++ {
		if (!ok || rand.Float64() >= 0.6) && rand.Intn(level+1) > 0 {
			after--
		}
	}
	return after
}

// CompatibleWithEnchantment ...
func (Unbreaking) CompatibleWithEnchantment(item.EnchantmentType) bool {
	return true
}

// CompatibleWithItem ...
func (Unbreaking) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Durable)
	return ok
}
