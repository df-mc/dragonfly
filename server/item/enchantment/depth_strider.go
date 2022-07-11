package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// DepthStrider is a boot enchantment that increases underwater movement speed.
type DepthStrider struct{}

// Name ...
func (DepthStrider) Name() string {
	return "Depth Strider"
}

// MaxLevel ...
func (DepthStrider) MaxLevel() int {
	return 3
}

// CompatibleWith ...
func (DepthStrider) CompatibleWith(s item.Stack) bool {
	b, ok := s.Item().(item.BootsType)
	return ok && b.Boots()
}
