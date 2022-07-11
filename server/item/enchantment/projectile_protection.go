package enchantment

import "github.com/df-mc/dragonfly/server/item"

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

// CompatibleWith ...
func (ProjectileProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	_, blastProt := s.Enchantment(BlastProtection{})
	_, fireProt := s.Enchantment(FireProtection{})
	_, prot := s.Enchantment(Protection{})
	return ok && !blastProt && !fireProt && !prot
}
