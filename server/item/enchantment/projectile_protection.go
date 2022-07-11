package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
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

// Affects ...
func (ProjectileProtection) Affects(src damage.Source) bool {
	_, projectile := src.(damage.SourceProjectile)
	return projectile
}

// Modifier returns the base protection modifier for the enchantment.
func (ProjectileProtection) Modifier() float64 {
	return 1.5
}

// CompatibleWith ...
func (ProjectileProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	// TODO: Ensure that the armour does not have blast protection.
	_, fireProt := s.Enchantment(FireProtection{})
	_, prot := s.Enchantment(Protection{})
	return ok && !fireProt && !prot
}
