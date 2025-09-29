package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"iter"
	"math"
	"time"
)

// FireworkBehaviourConfig holds optional parameters for a FireworkBehaviour.
type FireworkBehaviourConfig struct {
	Firework item.Firework
	Owner    *world.EntityHandle
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

func (conf FireworkBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates a FireworkBehaviour for an fw and owner using the optional
// parameters in conf.
func (conf FireworkBehaviourConfig) New() *FireworkBehaviour {
	b := &FireworkBehaviour{conf: conf}
	b.passive = PassiveBehaviourConfig{
		ExistenceDuration: conf.ExistenceDuration,
		Expire:            b.explode,
		Tick:              b.tick,
	}.New()
	return b
}

// FireworkBehaviour implements Behaviour for a firework entity.
type FireworkBehaviour struct {
	conf    FireworkBehaviourConfig
	passive *PassiveBehaviour
}

// Firework returns the underlying item.Firework of the FireworkBehaviour.
func (f *FireworkBehaviour) Firework() item.Firework {
	return f.conf.Firework
}

// Attached specifies if the firework is attached to its owner.
func (f *FireworkBehaviour) Attached() bool {
	return f.conf.Attached
}

// Owner returns the world.Entity that launched the firework.
func (f *FireworkBehaviour) Owner() *world.EntityHandle {
	return f.conf.Owner
}

// Tick moves the firework and makes it explode when it reaches its maximum
// duration.
func (f *FireworkBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	return f.passive.Tick(e, tx)
}

// tick ticks the entity, updating its velocity either with a constant factor
// or based on the owner's position and velocity if attached.
func (f *FireworkBehaviour) tick(e *Ent, tx *world.Tx) {
	owner, ok := f.conf.Owner.Entity(tx)
	if f.conf.Attached && ok {
		// The client will propel itself to match the firework's velocity since
		// we set the appropriate metadata.
		e.data.Pos = owner.Position()
	} else {
		e.data.Vel[0] *= f.conf.SidewaysVelocityMultiplier
		e.data.Vel[1] += f.conf.UpwardsAcceleration
		e.data.Vel[2] *= f.conf.SidewaysVelocityMultiplier
	}
}

// explode causes an explosion at the position of the firework, spawning
// particles and damaging nearby entities.
func (f *FireworkBehaviour) explode(e *Ent, tx *world.Tx) {
	owner, _ := f.conf.Owner.Entity(tx)
	pos, explosions := e.Position(), f.conf.Firework.Explosions

	for _, v := range tx.Viewers(pos) {
		v.ViewEntityAction(e, FireworkExplosionAction{})
	}
	for _, explosion := range explosions {
		if explosion.Shape == item.FireworkShapeHugeSphere() {
			tx.PlaySound(pos, sound.FireworkHugeBlast{})
		} else {
			tx.PlaySound(pos, sound.FireworkBlast{})
		}
		if explosion.Twinkle {
			tx.PlaySound(pos, sound.FireworkTwinkle{})
		}
	}

	if len(explosions) == 0 {
		return
	}

	force := float64(len(explosions)*2) + 5.0
	for victim := range filterLiving(tx.EntitiesWithin(e.H().Type().BBox(e).Translate(pos).Grow(5.25))) {
		tpos := victim.Position()
		dist := pos.Sub(tpos).Len()
		if dist > 5.0 {
			// The maximum distance allowed is 5.0 blocks.
			continue
		}
		dmg := force * math.Sqrt((5.0-dist)/5.0)
		src := ProjectileDamageSource{Owner: owner, Projectile: e}

		if pos == tpos {
			victim.(Living).Hurt(dmg, src)
			continue
		}
		if _, ok := trace.Perform(pos, tpos, tx, victim.H().Type().BBox(victim).Grow(0.3), nil); ok {
			victim.(Living).Hurt(dmg, src)
		}
	}
}

func filterLiving(seq iter.Seq[world.Entity]) iter.Seq[world.Entity] {
	return func(yield func(world.Entity) bool) {
		for e := range seq {
			if _, living := e.(Living); !living {
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
}
