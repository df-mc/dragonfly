package enchantment

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/item"
)

// FeatherFalling is an enchantment to boots that reduces fall damage. It does not affect falling speed.
type FeatherFalling struct{}

// Name ...
func (FeatherFalling) Name() string {
	return "Feather Falling"
}

// MaxLevel ...
func (FeatherFalling) MaxLevel() int {
	return 4
}

// Affects ...
func (FeatherFalling) Affects(src damage.Source) bool {
	_, fall := src.(damage.SourceFall)
	return fall
}

// Modifier returns the base protection modifier for the enchantment.
func (FeatherFalling) Modifier() float64 {
	return 2.5
}

// CompatibleWith ...
func (FeatherFalling) CompatibleWith(s item.Stack) bool {
	b, ok := s.Item().(item.BootsType)
	return ok && b.Boots()
}
