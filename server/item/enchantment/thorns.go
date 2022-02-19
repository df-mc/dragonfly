package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Thorns is an armor enchantment that causes attackers to be damaged when they deal damage to the wearer.
type Thorns struct{}

// Name ...
func (e Thorns) Name() string {
	return "Thorns"
}

// MaxLevel ...
func (e Thorns) MaxLevel() int {
	return 3
}

// CompatibleWith ...
func (e Thorns) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Armour)
	return ok
}
