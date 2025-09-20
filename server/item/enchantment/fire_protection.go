package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// FireProtection is an armour enchantment that decreases fire damage.
var FireProtection fireProtection

type fireProtection struct{}

func (fireProtection) Name() string {
	return "Fire Protection"
}

func (fireProtection) MaxLevel() int {
	return 4
}

func (fireProtection) Cost(level int) (int, int) {
	minCost := 10 + (level-1)*8
	return minCost, minCost + 8
}

func (fireProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Modifier returns the base protection modifier for the enchantment.
func (fireProtection) Modifier() float64 {
	return 0.08
}

func (fireProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != BlastProtection && t != ProjectileProtection && t != Protection
}

func (fireProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
