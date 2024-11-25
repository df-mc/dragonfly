package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// ProjectileProtection is an armour enchantment that reduces damage from
// projectiles.
var ProjectileProtection projectileProtection

type projectileProtection struct{}

// Name ...
func (projectileProtection) Name() string {
	return "Projectile Protection"
}

// MaxLevel ...
func (projectileProtection) MaxLevel() int {
	return 4
}

// Cost ...
func (projectileProtection) Cost(level int) (int, int) {
	minCost := 3 + (level-1)*6
	return minCost, minCost + 6
}

// Rarity ...
func (projectileProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Modifier returns the base protection modifier for the enchantment.
func (projectileProtection) Modifier() float64 {
	return 0.08
}

// CompatibleWithEnchantment ...
func (projectileProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return t != BlastProtection && t != FireProtection && t != Protection
}

// CompatibleWithItem ...
func (projectileProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
