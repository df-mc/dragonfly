package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// SwiftSneak is a non-renewable enchantment that can be applied to leggings and allows the player to walk more quickly
// while sneaking.
type SwiftSneak struct{}

// Name ...
func (SwiftSneak) Name() string {
	return "Swift Sneak"
}

// MaxLevel ...
func (SwiftSneak) MaxLevel() int {
	return 3
}

// CompatibleWith ...
func (SwiftSneak) CompatibleWith(s item.Stack) bool {
	b, ok := s.Item().(item.LeggingsType)
	return ok && b.Leggings()
}
