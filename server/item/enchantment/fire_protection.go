package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// FireProtection is an armour enchantment that decreases fire damage.
type FireProtection struct{}

// Name ...
func (FireProtection) Name() string {
	return "Fire Protection"
}

// MaxLevel ...
func (FireProtection) MaxLevel() int {
	return 4
}

// MinCost ...
func (FireProtection) MinCost(level int) int {
	return 10 + (level-1)*8
}

// MaxCost ...
func (f FireProtection) MaxCost(level int) int {
	return f.MinCost(level) + 8
}

// Rarity ...
func (FireProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// CompatibleWithOther ...
func (FireProtection) CompatibleWithOther(t item.EnchantmentType) bool {
	_, blastProt := t.(BlastProtection)
	_, projectileProt := t.(ProjectileProtection)
	_, prot := t.(Protection)
	return !blastProt && !projectileProt && !prot
}

// CompatibleWithItem ...
func (FireProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
