package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
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

// Rarity ...
func (ProjectileProtection) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityUncommon
}

// Affects ...
func (ProjectileProtection) Affects(src damage.Source) bool {
	_, projectile := src.(damage.SourceProjectile)
	return projectile
}

// Modifier returns the base protection modifier for the enchantment.
func (ProjectileProtection) Modifier() float64 {
	return 1.5
}

// CompatibleWithEnchantment ...
func (ProjectileProtection) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	// TODO: Ensure that the armour does not have blast protection.
	_, fireProtection := t.(FireProtection)
	_, protection := t.(Protection)
	return !fireProtection && !protection
}

// CompatibleWithItem ...
func (ProjectileProtection) CompatibleWithItem(i world.Item) bool {
	_, ok := i.(item.Armour)
	return ok
}
