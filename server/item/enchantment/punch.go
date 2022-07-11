package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// Punch increases the knock-back dealt when hitting a player or mob with a bow.
type Punch struct{}

// Name ...
func (Punch) Name() string {
	return "Punch"
}

// MaxLevel ...
func (Punch) MaxLevel() int {
	return 2
}

// Multiplier returns the knock-back multiplier for the level and horizontal speed.
func (Punch) Multiplier(level int, horizontalSpeed float64) float64 {
	return float64(level) * 0.6 / horizontalSpeed
}

// CompatibleWith ...
func (Punch) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Bow)
	return ok
}
