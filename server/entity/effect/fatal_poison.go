package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// FatalPoison is a lasting effect that causes the affected entity to lose
// health gradually. fatalPoison, unlike poison, can kill the entity it is
// applied to.
var FatalPoison fatalPoison

type fatalPoison struct {
	nopLasting
}

// Apply ...
func (fatalPoison) Apply(e world.Entity, eff Effect) {
	interval := max(50>>(eff.Level()-1), 1)
	if eff.Tick()%interval == 0 {
		if l, ok := e.(living); ok {
			l.Hurt(1, PoisonDamageSource{Fatal: true})
		}
	}
}

// RGBA ...
func (fatalPoison) RGBA() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x93, B: 0x31, A: 0xff}
}
