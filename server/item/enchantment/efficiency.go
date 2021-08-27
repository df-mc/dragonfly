package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// Efficiency is an enchantment that increases mining speed.
type Efficiency struct{ enchantment }

// Addend returns the mining speed addend from efficiency.
func (e Efficiency) Addend(level int) float64 {
	return float64(level*level + 1)
}

// Name ...
func (e Efficiency) Name() string {
	return "Efficiency"
}

// MaxLevel ...
func (e Efficiency) MaxLevel() int {
	return 5
}

// WithLevel ...
func (e Efficiency) WithLevel(level int) item.Enchantment {
	return Efficiency{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e Efficiency) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(tool.Tool)
	return ok && t.ToolType() == tool.TypePickaxe
}
