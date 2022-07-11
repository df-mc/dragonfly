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
	_, ok := src.(damage.SourceEntityAttack)
	return ok || src == damage.SourceFall{} || src == damage.SourceFire{} || src == damage.SourceFireTick{} || src == damage.SourceLava{}
}

// Multiplier returns the damage multiplier of protection.
func (Protection) Multiplier(lvl int) float64 {
	if lvl > 20 {
		lvl = 20
	}
	return 1 - float64(lvl)/25
}

// CompatibleWith ...
func (Protection) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	_, blastProt := s.Enchantment(BlastProtection{})
	_, fireProt := s.Enchantment(FireProtection{})
	_, projectileProt := s.Enchantment(ProjectileProtection{})
	return ok && !blastProt && !fireProt && !projectileProt
}
