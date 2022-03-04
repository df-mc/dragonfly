package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// FeatherFalling is an enchantment to boots that reduces fall damage. It does not affect falling speed.
type FeatherFalling struct{}

// Multiplier returns the damage multiplier of feather falling.
func (e FeatherFalling) Multiplier(lvl int) float64 {
	return 1 - 0.12*float64(lvl)
}

// Name ...
func (e FeatherFalling) Name() string {
	return "Feather Falling"
}

// MaxLevel ...
func (e FeatherFalling) MaxLevel() int {
	return 4
}

// CompatibleWith ...
func (e FeatherFalling) CompatibleWith(s item.Stack) bool {
	b, ok := s.Item().(item.BootsType)
	return ok && b.Boots()
}
