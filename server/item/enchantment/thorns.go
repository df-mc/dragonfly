package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Thorns is an enchantment that inflicts damage on attackers.
type Thorns struct{}

// Name ...
func (Thorns) Name() string {
	return "Thorns"
}

// MaxLevel ...
func (Thorns) MaxLevel() int {
	return 3
}

// CompatibleWith ...
func (Thorns) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	return ok
}
