package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// FeatherFalling is an enchantment to boots that reduces fall damage. It does not affect falling speed.
type FeatherFalling struct{}

// Multiplier returns the damage multiplier of feather falling.
func (FeatherFalling) Multiplier(lvl int) float64 {
	return 1 - 0.12*float64(lvl)
}

// Name ...
func (FeatherFalling) Name() string {
	return "Feather Falling"
}

// MaxLevel ...
func (FeatherFalling) MaxLevel() int {
	return 4
}

// CompatibleWith ...
func (FeatherFalling) CompatibleWith(s item.Stack) bool {
	b, ok := s.Item().(item.BootsType)
	return ok && b.Boots()
}
