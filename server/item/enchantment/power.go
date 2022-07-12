package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Power is a bow enchantment which increases arrow damage.
type Power struct{}

// Name ...
func (Power) Name() string {
	return "Power"
}

// MaxLevel ...
func (Power) MaxLevel() int {
	return 5
}

// PowerDamage returns the extra base damage dealt by the enchantment and level.
func (Power) PowerDamage(level int) float64 {
	return (float64(level) + 1) / 2
}

// CompatibleWith ...
func (Power) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Bow)
	return ok
}
