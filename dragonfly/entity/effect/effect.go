package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// instantEffect forms the base of an instant effect.
type instantEffect struct {
	// Lvl holds the level of the effect. A higher level results in a more powerful effect, whereas a negative
	// level will generally inverse effect.
	Lvl int
}

// Instant always returns true for instant effects.
func (instantEffect) Instant() bool {
	return true
}

// Level returns the level of the instant effect.
func (i instantEffect) Level() int {
	return i.Lvl
}

// Duration always returns 0 for instant effects.
func (instantEffect) Duration() time.Duration {
	return 0
}

// ShowParticles always returns false for instant effects.
func (instantEffect) ShowParticles() bool {
	return false
}

// AmbientSource always returns false for instant effects.
func (instantEffect) AmbientSource() bool {
	return false
}

// RGBA always returns an empty color.RGBA.
func (instantEffect) RGBA() color.RGBA {
	return color.RGBA{}
}

// End ...
func (instantEffect) End(entity.Living) {}

// Start ...
func (instantEffect) Start(entity.Living) {}

// lastingEffect forms the base of an effect that lasts for a specific duration.
type lastingEffect struct {
	// Lvl holds the level of the effect. A higher level results in a more powerful effect, whereas a negative
	// level will generally inverse effect.
	Lvl int
	// Dur holds the duration of the effect. One will be subtracted every time the entity that the effect is
	// added to is ticked.
	Dur time.Duration
	// HideParticles hides the coloured particles of the effect when added to an entity.
	HideParticles bool
	// Ambient specifies if the effect comes from an ambient source, such as from a beacon or conduit. The
	// particles displayed when Ambient is true are less visible.
	Ambient bool
}

// Instant always returns false for lasting effects.
func (lastingEffect) Instant() bool {
	return false
}

// Level returns the level of the lasting effect.
func (l lastingEffect) Level() int {
	return l.Lvl
}

// Duration returns the leftover duration of the lasting effect.
func (l lastingEffect) Duration() time.Duration {
	return l.Dur
}

// ShowParticles returns true if the effect does not display particles.
func (l lastingEffect) ShowParticles() bool {
	return !l.HideParticles
}

// AmbientSource specifies if the effect comes from a beacon or conduit.
func (l lastingEffect) AmbientSource() bool {
	return l.Ambient
}

// withSettings returns the lastingEffect with the duration passed.
func (l lastingEffect) withSettings(d time.Duration, level int, ambient bool) lastingEffect {
	l.Dur = d
	l.Lvl = level
	l.Ambient = ambient
	return l
}

// End ...
func (lastingEffect) End(entity.Living) {}

// Start ...
func (lastingEffect) Start(entity.Living) {}

// Apply ...
func (lastingEffect) Apply(living entity.Living) {}

// tickDuration returns the duration as in-game ticks.
func tickDuration(d time.Duration) int {
	return int(d / (time.Second / 20))
}

// ResultingColour calculates the resulting colour of the effects passed and returns a bool specifying if the
// effects were ambient effects, which will cause their particles to display less frequently.
func ResultingColour(effects []entity.Effect) (color.RGBA, bool) {
	r, g, b, a := 0, 0, 0, 0
	l := len(effects)
	if l == 0 {
		return color.RGBA{}, false
	}

	ambient := true
	for _, e := range effects {
		c := e.RGBA()
		r += int(c.R)
		g += int(c.G)
		b += int(c.B)
		a += int(c.A)
		if !e.AmbientSource() {
			ambient = false
		}
	}
	return color.RGBA{R: uint8(r / l), G: uint8(g / l), B: uint8(b / l), A: uint8(a / l)}, ambient
}
