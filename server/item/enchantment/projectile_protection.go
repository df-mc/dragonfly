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

// MinCost ...
func (ProjectileProtection) MinCost(level int) int {
	return 3 + (level-1)*6
}

// MaxCost ...
func (p ProjectileProtection) MaxCost(level int) int {
	return p.MinCost(level) + 6
}

// Rarity ...
func (ProjectileProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// CompatibleWithEnchantment ...
func (ProjectileProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	_, blastProt := t.(BlastProtection)
	_, fireProt := t.(FireProtection)
	_, prot := t.(Protection)
	return !blastProt && !fireProt && !prot
}

// CompatibleWithItem ...
func (ProjectileProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
