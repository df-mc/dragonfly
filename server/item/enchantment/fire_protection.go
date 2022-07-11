package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
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

// Affects ...
func (FireProtection) Affects(src damage.Source) bool {
	_, fire := src.(damage.SourceFire)
	_, fireTick := src.(damage.SourceFireTick)
	return fire || fireTick
}

// Modifier returns the base protection modifier for the enchantment.
func (FireProtection) Modifier() float64 {
	return 1.25
}

// CompatibleWith ...
func (FireProtection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	// TODO: Ensure that the armour does not have blast protection.
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	_, prot := s.Enchantment(Protection{})
	return ok && !projectileProt && !prot
}
