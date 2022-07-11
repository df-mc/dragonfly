package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// KnockBack is an enchantment to a sword that increases the sword's knock-back.
type KnockBack struct{}

// Name ...
func (e KnockBack) Name() string {
	return "Knockback"
}

// MaxLevel ...
func (e KnockBack) MaxLevel() int {
	return 2
}

// Force returns the increase in knock-back force from the enchantment.
func (e KnockBack) Force(lvl int) float64 {
	return float64(lvl) / 2
}

// CompatibleWith ...
func (e KnockBack) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
