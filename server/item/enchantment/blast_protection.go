package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// BlastProtection is an armour enchantment that decreases explosion damage.
type BlastProtection struct{}

// Name ...
func (BlastProtection) Name() string {
	return "Blast Protection"
}

// MaxLevel ...
func (BlastProtection) MaxLevel() int {
	return 4
}

// MinCost ...
func (BlastProtection) MinCost(level int) int {
	return 5 + (level-1)*8
}

// MaxCost ...
func (b BlastProtection) MaxCost(level int) int {
	return b.MinCost(level) + 8
}

// Rarity ...
func (BlastProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// CompatibleWithEnchantment ...
func (BlastProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, fireProt := t.(FireProtection)
	_, projectileProt := t.(ProjectileProtection)
	_, prot := t.(Protection)
	return !fireProt && !projectileProt && !prot
}

// CompatibleWithItem ...
func (BlastProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
