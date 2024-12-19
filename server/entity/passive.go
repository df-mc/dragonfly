package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
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
	Expire func(e *Ent, tx *world.Tx)
	// Tick is called for every tick that the entity is alive. Tick is called
	// after the entity moves on a tick.
	Tick func(e *Ent, tx *world.Tx)
}

func (conf PassiveBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates a PassiveBehaviour using the parameters in conf.
func (conf PassiveBehaviourConfig) New() *PassiveBehaviour {
	if conf.ExistenceDuration == 0 {
		conf.ExistenceDuration = math.MaxInt64
	}
	return &PassiveBehaviour{conf: conf, fuse: conf.ExistenceDuration, mc: &MovementComputer{
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

	close        bool
	fallDistance float64
	fuse         time.Duration
}

// Explode adds velocity to a passive entity to blast it away from the
// explosion's source.
func (p *PassiveBehaviour) Explode(e *Ent, src mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	e.data.Vel = e.data.Vel.Add(e.data.Pos.Sub(src).Normalize().Mul(impact))
}

// Fuse returns the leftover time until PassiveBehaviourConfig.Expire is called,
// or -1 if this function is not set.
func (p *PassiveBehaviour) Fuse() time.Duration {
	if p.conf.Expire != nil {
		return p.fuse
	}
	return -1
}

// Tick implements the behaviour for a passive entity. It performs movement and
// updates its state.
func (p *PassiveBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if p.close {
		_ = e.Close()
		return nil
	}

	m := p.mc.TickMovement(e, e.data.Pos, e.data.Vel, e.data.Rot, tx)
	e.data.Pos, e.data.Vel = m.pos, m.vel
	p.fallDistance = math.Max(p.fallDistance-m.dvel[1], 0)

	p.fuse = p.conf.ExistenceDuration - e.Age()

	if p.conf.Tick != nil {
		p.conf.Tick(e, tx)
	}

	if p.Fuse()%(time.Second/4) == 0 {
		for _, v := range tx.Viewers(m.pos) {
			v.ViewEntityState(e)
		}
	}

	if e.Age() > p.conf.ExistenceDuration {
		p.close = true
		if p.conf.Expire != nil {
			p.conf.Expire(e, tx)
		}
	}
	return m
}
