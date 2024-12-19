package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// FatalPoison is a lasting effect that causes the affected entity to lose health gradually. FatalPoison,
// unlike Poison, can kill the entity it is applied to.
type FatalPoison struct {
	nopLasting
}

// Apply ...
func (FatalPoison) Apply(e world.Entity, eff Effect) {
	interval := max(50>>(eff.Level()-1), 1)
	if eff.Tick()%interval == 0 {
		if l, ok := e.(living); ok {
			l.Hurt(1, PoisonDamageSource{Fatal: true})
		}
	}
}

// RGBA ...
func (FatalPoison) RGBA() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x93, B: 0x31, A: 0xff}
}
