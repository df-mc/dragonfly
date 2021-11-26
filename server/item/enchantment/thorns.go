package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/armour"
)

// Thorns is an armor enchantment that causes attackers to be damaged when they deal damage to the wearer.
type Thorns struct{ enchantment }

// Name ...
func (e Thorns) Name() string {
	return "Thorns"
}

// MaxLevel ...
func (e Thorns) MaxLevel() int {
	return 3
}

// WithLevel ...
func (e Thorns) WithLevel(level int) item.Enchantment {
	return Thorns{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e Thorns) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(armour.Armour)
	return ok
}
