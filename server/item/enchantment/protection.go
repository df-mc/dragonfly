package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
)

// Protection is an armour enchantment which increases the damage reduction.
type Protection struct{}

// Name ...
func (Protection) Name() string {
	return "Protection"
}

// MaxLevel ...
func (Protection) MaxLevel() int {
	return 4
}

// Affects ...
func (Protection) Affects(src damage.Source) bool {
	_, projectile := src.(damage.SourceProjectile)
	_, attack := src.(damage.SourceEntityAttack)
	_, fireTick := src.(damage.SourceFireTick)
	_, fall := src.(damage.SourceFall)
	_, fire := src.(damage.SourceFire)
	_, lava := src.(damage.SourceLava)
	return projectile || attack || fireTick || fall || fire || lava
}

// Modifier returns the base protection modifier for the enchantment.
func (Protection) Modifier() float64 {
	return 0.75
}

// CompatibleWith ...
func (Protection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	// TODO: Ensure that the armour does not have blast protection.
	_, fireProt := s.Enchantment(FireProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	return ok && !fireProt && !projectileProt
}
