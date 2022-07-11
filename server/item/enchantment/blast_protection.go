package enchantment

import "github.com/df-mc/dragonfly/server/item"

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

// CompatibleWith ...
func (BlastProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	_, fireProt := s.Enchantment(FireProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	_, prot := s.Enchantment(Protection{})
	return ok && !fireProt && !projectileProt && !prot
}
