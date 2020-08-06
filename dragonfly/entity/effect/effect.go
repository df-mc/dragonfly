package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Effect represents an effect that may be added to a living entity. Effects may either be instant or last
// for a specific duration.
type Effect interface {
	// Instant checks if the effect is instance. If it is instant, the effect will only be ticked a single
	// time when added to an entity.
	Instant() bool
	// Apply applies the effect to an entity. For instant effects, this method applies the effect once, such
	// as healing the world.Entity for instant health.
	Apply(e world.Entity)
	// Level returns the level of the effect. A higher level generally means a more powerful effect.
	Level() int
	// Duration returns the leftover duration of the effect.
	Duration() time.Duration
	// WithSettings returns the effect with a duration and level passed.
	WithSettings(d time.Duration, level int, ambient bool) Effect
	// RGBA returns the colour of the effect. If multiple effects are present, the colours will be mixed
	// together to form a new colour.
	RGBA() color.RGBA
	// ShowParticles checks if the particle should show particles. If not, entities that have the effect
	// will not display particles around them.
	ShowParticles() bool
	// AmbientSource specifies if the effect came from an ambient source, such as a beacon or conduit. The
	// particles will be less visible when this is true.
	AmbientSource() bool
	// Start is called for lasting events. It is sent the first time the effect is applied to an entity.
	Start(e world.Entity)
	// End is called for lasting events. It is sent the moment the effect expires.
	End(e world.Entity)
}

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
func (instantEffect) End(world.Entity) {}

// Start ...
func (instantEffect) Start(world.Entity) {}

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
func (lastingEffect) End(world.Entity) {}

// Start ...
func (lastingEffect) Start(world.Entity) {}

// Apply ...
func (lastingEffect) Apply(living world.Entity) {}

// tickDuration returns the duration as in-game ticks.
func tickDuration(d time.Duration) int {
	return int(d / (time.Second / 20))
}

// ResultingColour calculates the resulting colour of the effects passed and returns a bool specifying if the
// effects were ambient effects, which will cause their particles to display less frequently.
func ResultingColour(effects []Effect) (color.RGBA, bool) {
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

// living represents a living entity that has health and the ability to move around.
type living interface {
	world.Entity
	// Health returns the health of the entity.
	Health() float64
	// MaxHealth returns the maximum health of the entity.
	MaxHealth() float64
	// SetMaxHealth changes the maximum health of the entity to the value passed.
	SetMaxHealth(v float64)
	// Hurt hurts the entity for a given amount of damage. The source passed represents the cause of the
	// damage, for example damage.SourceEntityAttack if the entity is attacked by another entity.
	// If the final damage exceeds the health that the player currently has, the entity is killed.
	Hurt(damage float64, source damage.Source)
	// Heal heals the entity for a given amount of health. The source passed represents the cause of the
	// healing, for example healing.SourceFood if the entity healed by having a full food bar. If the health
	// added to the original health exceeds the entity's max health, Heal may not add the full amount.
	Heal(health float64, source healing.Source)
	// Speed returns the current speed of the living entity. The default value is different for each entity.
	Speed() float64
	// SetSpeed sets the speed of an entity to a new value.
	SetSpeed(float64)
}
