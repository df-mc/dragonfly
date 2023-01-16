package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// ProjectileProtection is an armour enchantment that reduces damage from projectiles.
type ProjectileProtection struct{}

// Name ...
func (ProjectileProtection) Name() string {
	return "Projectile Protection"
}

// MaxLevel ...
func (ProjectileProtection) MaxLevel() int {
	return 4
}

// Cost ...
func (ProjectileProtection) Cost(level int) (int, int) {
	min := 3 + (level-1)*6
	return min, min + 6
}

// Rarity ...
func (ProjectileProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Modifier returns the base protection modifier for the enchantment.
func (ProjectileProtection) Modifier() float64 {
	return 0.08
}

// CompatibleWithEnchantment ...
func (ProjectileProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, blastProtection := t.(BlastProtection)
	_, fireProtection := t.(FireProtection)
	_, protection := t.(Protection)
	return !blastProtection && !fireProtection && !protection
}

// CompatibleWithItem ...
func (ProjectileProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
