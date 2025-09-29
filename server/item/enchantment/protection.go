package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Protection is an armour enchantment which increases the damage reduction.
var Protection protection

type protection struct{}

func (protection) Name() string {
	return "Protection"
}

func (protection) MaxLevel() int {
	return 4
}

func (protection) Cost(level int) (int, int) {
	minCost := 1 + (level-1)*11
	return minCost, minCost + 11
}

func (protection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityCommon
}

// Modifier returns the base protection modifier for the enchantment.
func (protection) Modifier() float64 {
	return 0.04
}

func (protection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != BlastProtection && t != FireProtection && t != ProjectileProtection
}

func (protection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
