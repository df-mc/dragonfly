package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// SilkTouch is an enchantment that allows many blocks to drop themselves instead of their usual items when mined.
type SilkTouch struct{}

// Name ...
func (e SilkTouch) Name() string {
	return "Silk Touch"
}

// MaxLevel ...
func (e SilkTouch) MaxLevel() int {
	return 1
}

// CompatibleWith ...
func (e SilkTouch) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(item.Tool)
	//TODO: Fortune
	return ok && (t.ToolType() != item.TypeSword && t.ToolType() != item.TypeNone)
}
