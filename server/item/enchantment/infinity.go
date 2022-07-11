package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Infinity is an enchantment to bows that prevents regular arrows from being consumed when shot.
type Infinity struct{}

// Name ...
func (Infinity) Name() string {
	return "Infinity"
}

// MaxLevel ...
func (Infinity) MaxLevel() int {
	return 1
}

// CompatibleWith ...
func (Infinity) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Bow)
	return ok
}
