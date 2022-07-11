package enchantment

import "github.com/df-mc/dragonfly/server/item"

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

// CompatibleWith ...
func (FireProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	_, blastProt := s.Enchantment(BlastProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	_, prot := s.Enchantment(Protection{})
	return ok && !blastProt && !projectileProt && !prot
}
