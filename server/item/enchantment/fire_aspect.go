package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
	"time"
)

// FireAspect is a sword enchantment that sets the target on fire.
type FireAspect struct {
	enchantment
}

// Duration returns how long the fire from fire aspect will last.
func (e FireAspect) Duration(lvl int) time.Duration {
	return time.Second * 4 * time.Duration(lvl)
}

// Name ...
func (e FireAspect) Name() string {
	return "Fire Aspect"
}

// MaxLevel ...
func (e FireAspect) MaxLevel() int {
	return 2
}

// WithLevel ...
func (e FireAspect) WithLevel(level int) item.Enchantment {
	return FireAspect{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e FireAspect) CompatibleWith(s item.Stack) bool {
	t, ok := s.Item().(tool.Tool)
	return ok && t.ToolType() == tool.TypeSword
}
