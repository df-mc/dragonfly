package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// Sharpness is an enchantment applied to a sword or axe that increases melee damage.
type Sharpness struct {
	enchantment
}

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

// WithLevel ...
func (e Sharpness) WithLevel(level int) item.Enchantment {
	return Sharpness{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e Sharpness) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(tool.Tool)
	return ok && (t.ToolType() == tool.TypeSword || t.ToolType() == tool.TypeAxe)
}
