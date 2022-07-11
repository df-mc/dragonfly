package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// FireAspect is a sword enchantment that sets the target on fire.
type FireAspect struct{}

// Name ...
func (FireAspect) Name() string {
	return "Fire Aspect"
}

// MaxLevel ...
func (FireAspect) MaxLevel() int {
	return 2
}

// Duration returns how long the fire from fire aspect will last.
func (FireAspect) Duration(lvl int) time.Duration {
	return time.Second * 4 * time.Duration(lvl)
}

// CompatibleWith ...
func (FireAspect) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(item.Tool)
	return ok && t.ToolType() == item.TypeSword
}
