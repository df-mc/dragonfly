package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// SoulSpeed is an enchantment that can be applied on boots and allows the player to walk more quickly on soul sand or
// soul soil.
type SoulSpeed struct{}

// Name ...
func (SoulSpeed) Name() string {
	return "Soul Speed"
}

// MaxLevel ...
func (SoulSpeed) MaxLevel() int {
	return 3
}

// CompatibleWith ...
func (SoulSpeed) CompatibleWith(s item.Stack) bool {
	b, ok := s.Item().(item.BootsType)
	return ok && b.Boots()
}
