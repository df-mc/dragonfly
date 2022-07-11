package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// Flame turns your arrows into flaming arrows allowing you to set your targets on fire.
type Flame struct{}

// Name ...
func (Flame) Name() string {
	return "Flame"
}

// MaxLevel ...
func (Flame) MaxLevel() int {
	return 1
}

// Duration always returns a hundred seconds, no matter the level.
func (Flame) Duration() time.Duration {
	return time.Second * 100
}

// CompatibleWith ...
func (Flame) CompatibleWith(s item.Stack) bool {
	_, ok := s.Item().(item.Bow)
	return ok
}
