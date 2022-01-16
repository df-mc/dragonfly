package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
)

// AquaAffinity is a helmet enchantment that increases underwater mining speed.
type AquaAffinity struct {
	enchantment
}

// Name ...
func (e AquaAffinity) Name() string {
	return "Aqua Affinity"
}

// MaxLevel ...
func (e AquaAffinity) MaxLevel() int {
	return 1
}

// WithLevel ...
func (e AquaAffinity) WithLevel(level int) item.Enchantment {
	return AquaAffinity{e.withLevel(level, e)}
}

// CompatibleWith ...
func (e AquaAffinity) CompatibleWith(s item.Stack) bool {
	h, ok := s.Item().(item.HelmetType)
	return ok && h.Helmet()
}
