package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// BlastProtection is an armour enchantment that reduces damage from explosions.
type BlastProtection struct{}

// Name ...
func (BlastProtection) Name() string {
	return "Blast Protection"
}

// MaxLevel ...
func (BlastProtection) MaxLevel() int {
	return 4
}

// Cost ...
func (BlastProtection) Cost(level int) (int, int) {
	min := 5 + (level-1)*8
	return min, min + 8
}

// Rarity ...
func (BlastProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

// Modifier returns the base protection modifier for the enchantment.
func (BlastProtection) Modifier() float64 {
	return 0.08
}

// CompatibleWithEnchantment ...
func (BlastProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, fireProtection := t.(FireProtection)
	_, projectileProtection := t.(ProjectileProtection)
	_, protection := t.(Protection)
	return !fireProtection && !projectileProtection && !protection
}

// CompatibleWithItem ...
func (BlastProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
