package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Efficiency is an enchantment that increases mining speed.
type Efficiency struct{}

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

// CompatibleWith ...
func (e Efficiency) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(item.Tool)
	return ok && t.ToolType() == item.TypePickaxe
}
