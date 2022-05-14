package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Sharpness is an enchantment applied to a sword or axe that increases melee damage.
type Sharpness struct{}

// Addend returns the additional damage when attacking with sharpness.
func (e Sharpness) Addend(level int) float64 {
	return float64(level) * 1.25
}

// Name ...
func (e Sharpness) Name() string {
	return "Sharpness"
}

// MaxLevel ...
func (e Sharpness) MaxLevel() int {
	return 4
}

// CompatibleWith ...
func (e Sharpness) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(item.Tool)
	return ok && (t.ToolType() == item.TypeSword || t.ToolType() == item.TypeAxe)
}
