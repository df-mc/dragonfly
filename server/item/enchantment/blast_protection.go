package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// BlastProtection is an armour enchantment that reduces damage from explosions.
var BlastProtection blastProtection

type blastProtection struct{}

func (blastProtection) Name() string {
	return "Blast Protection"
}

func (blastProtection) MaxLevel() int {
	return 4
}

func (blastProtection) Cost(level int) (int, int) {
	minCost := 5 + (level-1)*8
	return minCost, minCost + 8
}

func (blastProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// Modifier returns the base protection modifier for the enchantment.
func (blastProtection) Modifier() float64 {
	return 0.08
}

func (blastProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != FireProtection && t != ProjectileProtection && t != Protection
}

func (blastProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
