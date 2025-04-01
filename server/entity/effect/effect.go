package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// LastingType represents an effect type that can have a duration. An effect
// can be made using it by calling effect.New with the LastingType.
type LastingType interface {
	Type
	// Start is called for lasting effects when they are initially added.
	Start(e world.Entity, lvl int)
	// End is called for lasting effects when they are removed.
	End(e world.Entity, lvl int)
}

// Type is an effect implementation that can be applied to an entity.
type Type interface {
	// RGBA returns the colour of the effect. If multiple effects are present,
	// the colours will be mixed together to form a new colour.
	RGBA() color.RGBA
	// Apply applies the effect to an entity. Apply is called only once for
	// instant effects, such as instantHealth, while it is called every tick for
	// lasting effects. The Effect holding the Type is passed along with the
	// current tick.
	Apply(e world.Entity, eff Effect)
}

// Effect is an effect that can be added to an entity. Effects are either
// instant (applying the effect only once) or lasting (applying the effect
// every tick).
type Effect struct {
	t                        Type
	d                        time.Duration
	lvl                      int
	potency                  float64
	ambient, particlesHidden bool
	infinite                 bool
	tick                     int
}

// NewInstant returns a new instant Effect using the Type passed. The effect
// will be applied to an entity once and will expire immediately after.
// NewInstant creates an Effect with a potency of 1.0.
func NewInstant(t Type, lvl int) Effect {
	return NewInstantWithPotency(t, lvl, 1)
}

// NewInstantWithPotency returns a new instant Effect using the Type and level
// passed. The effect will be applied to an entity once and expire immediately
// after. The potency passed additionally influences the strength of the effect.
// A higher potency (> 1.0) increases the effect power, while a lower potency
// (< 1.0) reduces it.
func NewInstantWithPotency(t Type, lvl int, potency float64) Effect {
	return Effect{t: t, lvl: lvl, potency: potency}
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

// NewInfinite creates a new Effect using a LastingType passed. Once added to an entity, the effect will persist indefinitely,
// until the effect is removed.
func NewInfinite(t LastingType, lvl int) Effect {
	return Effect{t: t, lvl: lvl, infinite: true}
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

// Duration returns the leftover duration of the Effect. The duration returned is always 0 if NewInstant or NewInfinite
// were used to create the effect.
func (e Effect) Duration() time.Duration {
	return e.d
}

// Ambient returns whether the Effect is an ambient effect, leading to reduced particles shown to the client. False is
// always returned if the Effect was created using New or NewInstant.
func (e Effect) Ambient() bool {
	return e.ambient
}

// Infinite returns if the Effect duration is infinite.
func (e Effect) Infinite() bool {
	return e.infinite
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
		if !e.Infinite() {
			e.d -= time.Second / 20
		}
		e.tick++
	}
	return e
}

// Tick returns the current tick of the Effect. This is the number of ticks that
// the Effect has been applied for.
func (e Effect) Tick() int {
	return e.tick
}

// nopLasting is a lasting effect with no (server-side) behaviour. It does not implement the RGBA method.
type nopLasting struct{}

func (nopLasting) Apply(world.Entity, Effect) {}
func (nopLasting) End(world.Entity, int)      {}
func (nopLasting) Start(world.Entity, int)    {}

// ResultingColour calculates the resulting colour of the effects passed and returns a bool specifying if the
// effects were ambient effects, which will cause their particles to display less frequently.
func ResultingColour(effects []Effect) (color.RGBA, bool) {
	r, g, b, a, n := 0, 0, 0, 0, 0
	ambient := true
	for _, e := range effects {
		if e.particlesHidden {
			// Don't take effects with hidden particles into account for colour
			// calculation: Their particles are hidden after all.
			continue
		}
		c := e.Type().RGBA()
		r += int(c.R)
		g += int(c.G)
		b += int(c.B)
		a += int(c.A)
		n++
		if !e.Ambient() {
			ambient = false
		}
	}
	if n == 0 {
		return color.RGBA{R: 0x38, G: 0x5d, B: 0xc6, A: 0xff}, false
	}
	return color.RGBA{R: uint8(r / n), G: uint8(g / n), B: uint8(b / n), A: uint8(a / n)}, ambient
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
	// damage, for example entity.AttackDamageSource if the entity is attacked by another entity.
	// If the final damage exceeds the health that the player currently has, the entity is killed.
	Hurt(damage float64, source world.DamageSource) (n float64, vulnerable bool)
	// Heal heals the entity for a given amount of health. The source passed represents the cause of the
	// healing, for example entity.FoodHealingSource if the entity healed by having a full food bar. If the health
	// added to the original health exceeds the entity's max health, Heal may not add the full amount.
	Heal(health float64, source world.HealingSource)
	// speed returns the current speed of the living entity. The default value is different for each entity.
	Speed() float64
	// SetSpeed sets the speed of an entity to a new value.
	SetSpeed(float64)
}
