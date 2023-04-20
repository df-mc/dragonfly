package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
)

// PassiveBehaviourConfig holds optional parameters for a PassiveBehaviour.
type PassiveBehaviourConfig struct {
	// Gravity is the amount of Y velocity subtracted every tick.
	Gravity float64
	// Drag is used to reduce all axes of the velocity every tick. Velocity is
	// multiplied with (1-Drag) every tick.
	Drag float64
	// ExistenceDuration is the duration that an entity with this behaviour
	// should last. Once this time expires, the entity is closed. If
	// ExistenceDuration is 0, the entity will never expire automatically.
	ExistenceDuration time.Duration
	// Expire is called when the entity expires due to its age reaching the
	// ExistenceDuration.
	Expire func(e *Ent)
}

// New creates a PassiveBehaviour using the parameters in conf.
func (conf PassiveBehaviourConfig) New() *PassiveBehaviour {
	if conf.ExistenceDuration == 0 {
		conf.ExistenceDuration = math.MaxInt64
	}
	return &PassiveBehaviour{conf: conf, mc: &MovementComputer{
		Gravity:           conf.Gravity,
		Drag:              conf.Drag,
		DragBeforeGravity: true,
	}}
}

// PassiveBehaviour implements Behaviour for entities that act passively. This
// means that they can move, but only under influence of the environment, which
// includes, for example, falling, and flowing water.
type PassiveBehaviour struct {
	conf PassiveBehaviourConfig
	mc   *MovementComputer

	close bool

	age          time.Duration
	fallDistance float64
}

// Explode adds velocity to a passive entity to blast it away from the
// explosion's source.
func (p *PassiveBehaviour) Explode(e *Ent, src mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	e.vel = e.vel.Add(e.pos.Sub(src).Normalize().Mul(impact))
}

// Fuse returns the leftover time until PassiveBehaviourConfig.Expire is called,
// or -1 if this function is not set.
func (p *PassiveBehaviour) Fuse() time.Duration {
	if p.conf.Expire != nil {
		return p.conf.ExistenceDuration - p.age
	}
	return -1
}

// Tick implements the behaviour for a passive entity. It performs movement and
// updates its state.
func (p *PassiveBehaviour) Tick(e *Ent) *Movement {
	if p.close {
		_ = e.Close()
		return nil
	}

	if p.age > p.conf.ExistenceDuration {
		p.close = true

		if p.conf.Expire != nil {
			p.conf.Expire(e)
		}
	}
	e.mu.Lock()
	p.age += time.Second / 20

	m := p.mc.TickMovement(e, e.pos, e.vel, 0, 0)
	e.pos, e.vel = m.pos, m.vel

	p.fallDistance = math.Max(p.fallDistance-m.dvel[1], 0)
	e.mu.Unlock()

	w := e.World()
	if p.Fuse()%5 == 0 {
		for _, v := range w.Viewers(m.pos) {
			v.ViewEntityState(e)
		}
	}
	return m
}
