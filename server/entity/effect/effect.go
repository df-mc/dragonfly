package effect

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/entity/healing"
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// LastingType represents an effect type that can have a duration. An effect can be made using it by calling effect.New
// with the LastingType.
type LastingType interface {
	Type
	// Start is called for lasting effects when they are initially added to an entity.
	Start(e world.Entity, lvl int)
	// End is called for lasting effects when they are removed from an entity.
	End(e world.Entity, lvl int)
}

// PotentType represents an effect type which can have its potency changed.
type PotentType interface {
	Type
	// WithPotency updates the potency of the type with the one given and returns it.
	WithPotency(potency float64) Type
}

// Type is an effect implementation that can be applied to an entity.
type Type interface {
	// RGBA returns the colour of the effect. If multiple effects are present, the colours will be mixed
	// together to form a new colour.
	RGBA() color.RGBA
	// Apply applies the effect to an entity. This method applies the effect to an entity once for instant effects, such
	// as healing the world.Entity for instant health.
	// Apply always has a duration of 0 passed to it for instant effect implementations. For lasting effects that
	// implement LastingType, the appropriate leftover duration is passed.
	Apply(e world.Entity, lvl int, d time.Duration)
}

// Effect is an effect that can be added to an entity. Effects are either instant (applying the effect only once) or
// lasting (applying the effect every tick).
type Effect struct {
	t                        Type
	d                        time.Duration
	lvl                      int
	ambient, particlesHidden bool
}

// NewInstant returns a new instant Effect using the Type passed. The effect will be applied to an entity once
// and will expire immediately after.
func NewInstant(t Type, lvl int) Effect {
	return Effect{t: t, lvl: lvl}
}

// New creates a new Effect using a LastingType passed. Once added to an entity, the time.Duration passed will be ticked down
// by the entity until it reaches a duration of 0.
func New(t LastingType, lvl int, d time.Duration) Effect {
	return Effect{t: t, lvl: lvl, d: d}
}

// NewAmbient creates a new ambient (reduced particles, as when using a beacon) Effect using a LastingType passed. Once added
// to an entity, the time.Duration passed will be ticked down by the entity until it reaches a duration of 0.
func NewAmbient(t LastingType, lvl int, d time.Duration) Effect {
	return Effect{t: t, lvl: lvl, d: d, ambient: true}
}

// WithoutParticles returns the same Effect with particles disabled. Adding the effect to players will not display the
// particles around the player.
func (e Effect) WithoutParticles() Effect {
	e.particlesHidden = true
	return e
}

// ParticlesHidden returns true if the Effect had its particles hidden by calling WithoutParticles.
func (e Effect) ParticlesHidden() bool {
	return e.particlesHidden
}

// Level returns the level of the Effect.
func (e Effect) Level() int {
	return e.lvl
}

// Duration returns the leftover duration of the Effect. The duration returned is always 0 if NewInstant was used to
// create the effect.
func (e Effect) Duration() time.Duration {
	return e.d
}

// Ambient returns whether the Effect is an ambient effect, leading to reduced particles shown to the client. False is
// always returned if the Effect was created using New or NewInstant.
func (e Effect) Ambient() bool {
	return e.ambient
}

// Type returns the underlying type of the Effect. It is either of the type Type or LastingType, depending on whether it
// was created using New or NewAmbient, or NewInstant.
func (e Effect) Type() Type {
	return e.t
}

// TickDuration ticks the effect duration, subtracting time.Second/20 from the leftover time and returning the resulting
// Effect.
func (e Effect) TickDuration() Effect {
	if _, ok := e.t.(LastingType); ok {
		e.d -= time.Second / 20
	}
	return e
}

// nopLasting is a lasting effect with no (server-side) behaviour. It does not implement the RGBA method.
type nopLasting struct{}

func (nopLasting) Apply(world.Entity, int, time.Duration) {}
func (nopLasting) End(world.Entity, int)                  {}
func (nopLasting) Start(world.Entity, int)                {}

// tickDuration returns the duration as in-game ticks.
func tickDuration(d time.Duration) int {
	return int(d / (time.Second / 20))
}

// ResultingColour calculates the resulting colour of the effects passed and returns a bool specifying if the
// effects were ambient effects, which will cause their particles to display less frequently.
func ResultingColour(effects []Effect) (color.RGBA, bool) {
	r, g, b, a, l := 0, 0, 0, 0, 0
	ambient := true
	for _, e := range effects {
		if e.particlesHidden {
			// Don't take effects with hidden particles into account for colour calculation: Their particles are hidden
			// after all.
			continue
		}
		c := e.Type().RGBA()
		r += int(c.R)
		g += int(c.G)
		b += int(c.B)
		a += int(c.A)
		l++
		if !e.Ambient() {
			ambient = false
		}
	}
	if l == 0 {
		// Prevent division by 0 errors if no effects with particles were present.
		return color.RGBA{}, false
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
	Hurt(damage float64, source damage.Source) (n float64, vulnerable bool)
	// Heal heals the entity for a given amount of health. The source passed represents the cause of the
	// healing, for example healing.SourceFood if the entity healed by having a full food bar. If the health
	// added to the original health exceeds the entity's max health, Heal may not add the full amount.
	Heal(health float64, source healing.Source)
	// Speed returns the current speed of the living entity. The default value is different for each entity.
	Speed() float64
	// SetSpeed sets the speed of an entity to a new value.
	SetSpeed(float64)
}
