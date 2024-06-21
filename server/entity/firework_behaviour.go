package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
)

// FireworkBehaviourConfig holds optional parameters for a FireworkBehaviour.
type FireworkBehaviourConfig struct {
	// ExistenceDuration is the duration that an entity with this behaviour
	// should last. Once this time expires, the entity is closed. If
	// ExistenceDuration is 0, the entity will never expire automatically.
	ExistenceDuration time.Duration
	// SidewaysVelocityMultiplier is a value that the sideways velocity (X/Z) is
	// multiplied with every tick. For normal fireworks this is 1.15.
	SidewaysVelocityMultiplier float64
	// UpwardsAcceleration is a value added to the firework's velocity every
	// tick. For normal fireworks, this is 0.04.
	UpwardsAcceleration float64
	// Attached specifies if the firework is attached to its owner. If true,
	// the firework will boost the speed of the owner while flying.
	Attached bool
}

// New creates a FireworkBehaviour for an fw and owner using the optional
// parameters in conf.
func (conf FireworkBehaviourConfig) New(fw item.Firework, owner world.Entity) *FireworkBehaviour {
	b := &FireworkBehaviour{conf: conf, firework: fw, owner: owner}
	b.passive = PassiveBehaviourConfig{
		ExistenceDuration: conf.ExistenceDuration,
		Expire:            b.explode,
		Tick:              b.tick,
	}.New()
	return b
}

// FireworkBehaviour implements Behaviour for a firework entity.
type FireworkBehaviour struct {
	conf FireworkBehaviourConfig

	passive *PassiveBehaviour

	firework item.Firework
	owner    world.Entity
}

// Firework returns the underlying item.Firework of the FireworkBehaviour.
func (f *FireworkBehaviour) Firework() item.Firework {
	return f.firework
}

// Attached specifies if the firework is attached to its owner.
func (f *FireworkBehaviour) Attached() bool {
	return f.conf.Attached
}

// Owner returns the world.Entity that launched the firework.
func (f *FireworkBehaviour) Owner() world.Entity {
	return f.owner
}

// Tick moves the firework and makes it explode when it reaches its maximum
// duration.
func (f *FireworkBehaviour) Tick(e *Ent) *Movement {
	return f.passive.Tick(e)
}

// tick ticks the entity, updating its velocity either with a constant factor
// or based on the owner's position and velocity if attached.
func (f *FireworkBehaviour) tick(e *Ent) {
	var ownerVel mgl64.Vec3
	if o, ok := f.owner.(interface {
		Velocity() mgl64.Vec3
	}); ok {
		ownerVel = o.Velocity()
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if f.conf.Attached {
		dV := f.owner.Rotation().Vec3()

		// The client will propel itself to match the firework's velocity since
		// we set the appropriate metadata.
		e.pos = f.owner.Position()
		e.vel = e.vel.Add(ownerVel.Add(dV.Mul(0.1).Add(dV.Mul(1.5).Sub(ownerVel).Mul(0.5))))
	} else {
		e.vel[0] *= f.conf.SidewaysVelocityMultiplier
		e.vel[1] += f.conf.UpwardsAcceleration
		e.vel[2] *= f.conf.SidewaysVelocityMultiplier
	}
}

// explode causes an explosion at the position of the firework, spawning
// particles and damaging nearby entities.
func (f *FireworkBehaviour) explode(e *Ent) {
	w, pos, explosions := e.World(), e.Position(), f.firework.Explosions

	for _, v := range w.Viewers(pos) {
		v.ViewEntityAction(e, FireworkExplosionAction{})
	}
	for _, explosion := range explosions {
		if explosion.Shape == item.FireworkShapeHugeSphere() {
			w.PlaySound(pos, sound.FireworkHugeBlast{})
		} else {
			w.PlaySound(pos, sound.FireworkBlast{})
		}
		if explosion.Twinkle {
			w.PlaySound(pos, sound.FireworkTwinkle{})
		}
	}

	if len(explosions) == 0 {
		return
	}

	force := float64(len(explosions)*2) + 5.0
	targets := w.EntitiesWithin(e.Type().BBox(e).Translate(pos).Grow(5.25), func(e world.Entity) bool {
		l, living := e.(Living)
		return !living || l.AttackImmune()
	})
	for _, e := range targets {
		tpos := e.Position()
		dist := pos.Sub(tpos).Len()
		if dist > 5.0 {
			// The maximum distance allowed is 5.0 blocks.
			continue
		}
		dmg := force * math.Sqrt((5.0-dist)/5.0)
		src := ProjectileDamageSource{Owner: f.owner, Projectile: e}

		if pos == tpos {
			e.(Living).Hurt(dmg, src)
			continue
		}
		if _, ok := trace.Perform(pos, tpos, w, e.Type().BBox(e).Grow(0.3), func(world.Entity) bool {
			return true
		}); ok {
			e.(Living).Hurt(dmg, src)
		}
	}
}
