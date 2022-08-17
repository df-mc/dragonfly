package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"iter"
	"math"
	"math/rand/v2"
	"time"
)

// ProjectileBehaviourConfig allows the configuration of projectiles. Calling
// ProjectileBehaviourConfig.New() creates a ProjectileBehaviour using these
// settings.
type ProjectileBehaviourConfig struct {
	Owner *world.EntityHandle
	// Gravity is the amount of Y velocity subtracted every tick.
	Gravity float64
	// Drag is used to reduce all axes of the velocity every tick. Velocity is
	// multiplied with (1-Drag) every tick.
	Drag float64
	// Damage specifies the base damage dealt by the Projectile. If set to a
	// negative number, entities hit are not hurt at all and are not knocked
	// back. The base damage is multiplied with the velocity of the projectile
	// to calculate the final damage of the projectile.
	Damage float64
	// Potion is the potion effect that is applied to an entity when the
	// projectile hits it.
	Potion potion.Potion
	// KnockBackForceAddend is the additional horizontal velocity that is
	// applied to an entity when it is hit by the projectile.
	KnockBackForceAddend float64
	// KnockBackHeightAddend is the additional vertical velocity that is applied
	// to an entity when it is hit by the projectile.
	KnockBackHeightAddend float64
	// Particle is a particle that is spawned when the projectile hits a
	// target, either a block or an entity. No particle is spawned if left nil.
	Particle world.Particle
	// ParticleCount is the amount of particles that should be spawned if
	// Particle is not nil. ParticleCount will be set to 1 if Particle is not
	// nil and ParticleCount is 0.
	ParticleCount int
	// Sound is a sound that is played when the projectile hits a target, either
	// a block or an entity. No sound is played if left nil.
	Sound world.Sound
	// Critical specifies if the projectile is critical. This spawns critical
	// hit particles behind the projectile and causes it to deal up to 50% more
	// damage.
	Critical bool
	// Hit is a function that is called when the projectile Ent hits a target
	// (the trace.Result). The target is either of the type trace.EntityResult
	// or trace.BlockResult. Hit may be set to run additional behaviour when a
	// projectile hits a target.
	Hit func(e *Ent, tx *world.Tx, target trace.Result)
	// SurviveBlockCollision specifies if a projectile with this
	// ProjectileBehaviour should survive collision with a block. If set to
	// false, the projectile will break when hitting a block (like a snowball).
	// If set to true, the projectile will survive like an arrow does.
	SurviveBlockCollision bool
	// BlockCollisionVelocityMultiplier is the multiplier used to modify the
	// velocity of a projectile that has SurviveBlockCollision set to true. The
	// default, 0, will cause the projectile to lose its velocity completely. A
	// multiplier such as 0.5 will reduce the projectile's velocity, but retain
	// half of it after inverting the axis on which the projectile collided.
	BlockCollisionVelocityMultiplier float64
	// DisablePickup specifies if picking up the projectile should be disabled,
	// which is relevant in the case SurviveBlockCollision is set to true. Some
	// projectiles, such as arrows, cannot be picked up if they are shot by
	// monsters like skeletons.
	DisablePickup bool
	// PickupItem is the item that is given to a player when it picks up this
	// projectile. If left as an empty item.Stack, no item is given upon pickup.
	PickupItem item.Stack
	// CollisionPosition specifies the position that the projectile is stuck
	// in. If non-empty, the entity will not move.
	CollisionPosition cube.Pos
}

func (conf ProjectileBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates a new ProjectileBehaviour using conf. The owner passed may be nil
// if the projectile does not have one.
func (conf ProjectileBehaviourConfig) New() *ProjectileBehaviour {
	if conf.ParticleCount == 0 && conf.Particle != nil {
		conf.ParticleCount = 1
	}
	return &ProjectileBehaviour{conf: conf, collided: conf.CollisionPosition != cube.Pos{}, collisionPos: conf.CollisionPosition, mc: &MovementComputer{
		Gravity:           conf.Gravity,
		Drag:              conf.Drag,
		DragBeforeGravity: true,
	}}
}

// ProjectileBehaviour implements the behaviour of projectiles. Its specifics
// may be configured using ProjectileBehaviourConfig.
type ProjectileBehaviour struct {
	conf        ProjectileBehaviourConfig
	mc          *MovementComputer
	ageCollided int
	close       bool

	collisionPos cube.Pos
	collided     bool
}

// Owner returns the owner of the projectile.
func (lt *ProjectileBehaviour) Owner() *world.EntityHandle {
	return lt.conf.Owner
}

// Explode adds velocity to a projectile to blast it away from the explosion's
// source.
func (lt *ProjectileBehaviour) Explode(e *Ent, src mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	e.data.Vel = e.Velocity().Add(e.Position().Sub(src).Normalize().Mul(impact))
}

// Potion returns the potion.Potion that is applied to an entity if hit by the
// projectile.
func (lt *ProjectileBehaviour) Potion() potion.Potion {
	return lt.conf.Potion
}

// Critical returns true if ProjectileBehaviourConfig.Critical was set to true
// and if the projectile has not collided.
func (lt *ProjectileBehaviour) Critical() bool {
	return lt.conf.Critical && !lt.collided
}

// Tick runs the tick-based behaviour of a ProjectileBehaviour and returns the
// Movement within the tick. Tick handles the movement, collision and hitting
// of a projectile.
func (lt *ProjectileBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	if lt.close {
		_ = e.Close()
		return nil
	}

	if lt.collided && lt.tickAttached(e, tx) {
		if lt.ageCollided > 1200 {
			lt.close = true
		}
		return nil
	}
	vel := e.Velocity()
	m, result := lt.tickMovement(e, tx)
	e.data.Pos, e.data.Vel = m.pos, m.vel

	lt.collisionPos, lt.collided, lt.ageCollided = cube.Pos{}, false, 0

	if result == nil {
		return m
	}

	for i := 0; i < lt.conf.ParticleCount; i++ {
		tx.AddParticle(result.Position(), lt.conf.Particle)
	}
	if lt.conf.Sound != nil {
		tx.PlaySound(result.Position(), lt.conf.Sound)
	}

	switch r := result.(type) {
	case trace.EntityResult:
		if l, ok := r.Entity().(Living); ok && lt.conf.Damage >= 0 {
			lt.hitEntity(l, e, vel)
		}
	case trace.BlockResult:
		bpos := r.BlockPosition()
		if t, ok := tx.Block(bpos).(block.TNT); ok && e.OnFireDuration() > 0 {
			t.Ignite(bpos, tx, nil)
		}
		if h, ok := tx.Block(bpos).(block.ProjectileHitter); ok {
			h.ProjectileHit(bpos, tx, e, r.Face())
		}
		if lt.conf.SurviveBlockCollision {
			lt.hitBlockSurviving(e, r, m, tx)
			return m
		}
	}
	if lt.conf.Hit != nil {
		lt.conf.Hit(e, tx, result)
	}

	lt.close = true
	return m
}

// tickAttached performs the attached logic for a projectile. It checks if the
// projectile is still attached to a block and if it can be picked up.
func (lt *ProjectileBehaviour) tickAttached(e *Ent, tx *world.Tx) bool {
	boxes := tx.Block(lt.collisionPos).Model().BBox(lt.collisionPos, tx)
	box := e.H().Type().BBox(e).Translate(e.Position())

	for _, bb := range boxes {
		if box.IntersectsWith(bb.Translate(lt.collisionPos.Vec3()).Grow(0.05)) {
			if lt.ageCollided > 5 && !lt.conf.DisablePickup {
				lt.tryPickup(e, tx)
			}
			lt.ageCollided++
			return true
		}
	}
	return false
}

// tryPickup checks for nearby projectile collectors and closes the entity if
// one was found.
func (lt *ProjectileBehaviour) tryPickup(e *Ent, tx *world.Tx) {
	translated := e.H().Type().BBox(e).Translate(e.Position())
	grown := translated.GrowVec3(mgl64.Vec3{1, 0.5, 1})
	for other := range tx.EntitiesWithin(translated.Grow(2)) {
		if !other.H().Type().BBox(other).Translate(other.Position()).IntersectsWith(grown) {
			continue
		}
		collector, ok := other.(Collector)
		if !ok {
			continue
		}
		if _, ok := collector.Collect(lt.conf.PickupItem); !ok {
			continue
		}

		// A collector was within range and able to pick up the entity.
		lt.close = true
		for _, viewer := range tx.Viewers(e.Position()) {
			viewer.ViewEntityAction(e, PickedUpAction{Collector: collector})
		}
	}
}

// hitBlockSurviving is called if
// ProjectileBehaviourConfig.SurviveBlockCollision is set to true and the
// projectile collides with a block. If the resulting velocity is roughly 0,
// it sets the projectile as having collided with the block.
func (lt *ProjectileBehaviour) hitBlockSurviving(e *Ent, r trace.BlockResult, m *Movement, tx *world.Tx) {
	// Create an epsilon for deciding if the projectile has slowed down enough
	// for us to consider it as having collided for the final time. We take the
	// square root because FloatEqualThreshold squares it, which is not what we
	// want.
	eps := math.Sqrt(0.1 * (1 - lt.conf.BlockCollisionVelocityMultiplier))
	if mgl64.FloatEqualThreshold(e.Velocity().Len(), 0, eps) {
		e.SetVelocity(mgl64.Vec3{})
		lt.collisionPos, lt.collided = r.BlockPosition(), true

		for _, v := range tx.Viewers(m.pos) {
			v.ViewEntityAction(e, ArrowShakeAction{Duration: time.Millisecond * 350})
			v.ViewEntityState(e)
		}
		return
	}
}

// hitEntity is called when a projectile hits a Living. It deals damage to the
// entity and knocks it back. Additionally, it applies any potion effects and
// fire if applicable.
func (lt *ProjectileBehaviour) hitEntity(l Living, e *Ent, vel mgl64.Vec3) {
	owner, _ := lt.conf.Owner.Entity(e.tx)
	src := ProjectileDamageSource{Projectile: e, Owner: owner}
	dmg := math.Ceil(lt.conf.Damage * vel.Len())
	if lt.conf.Critical {
		dmg += rand.Float64() * dmg / 2
	}
	if _, vulnerable := l.Hurt(lt.conf.Damage, src); vulnerable {
		l.KnockBack(l.Position().Sub(vel), 0.45+lt.conf.KnockBackForceAddend, 0.3608+lt.conf.KnockBackHeightAddend)

		for _, eff := range lt.conf.Potion.Effects() {
			l.AddEffect(eff)
		}
		if flammable, ok := l.(Flammable); ok && e.OnFireDuration() > 0 {
			flammable.SetOnFire(time.Second * 5)
		}
	}
}

// tickMovement ticks the movement of a projectile. It updates the position and
// rotation of the projectile based on its velocity and updates the velocity
// based on gravity and drag.
func (lt *ProjectileBehaviour) tickMovement(e *Ent, tx *world.Tx) (*Movement, trace.Result) {
	pos, vel := e.Position(), e.Velocity()
	viewers := tx.Viewers(pos)

	velBefore := vel
	vel = lt.mc.applyHorizontalForces(tx, pos, lt.mc.applyVerticalForces(vel))
	rot := cube.Rotation{
		mgl64.RadToDeg(math.Atan2(vel[0], vel[2])),
		mgl64.RadToDeg(math.Atan2(vel[1], math.Hypot(vel[0], vel[2]))),
	}

	var (
		end = pos.Add(vel)
		hit trace.Result
		ok  bool
	)
	if !mgl64.FloatEqual(end.Sub(pos).LenSqr(), 0) {
		if hit, ok = trace.Perform(pos, end, tx, e.H().Type().BBox(e).Grow(1.0), lt.ignores(e)); ok {
			if _, ok := hit.(trace.BlockResult); ok {
				// Undo the gravity because the velocity as a result of gravity
				// at the point of collision should be 0.
				vel[1] = (vel[1] + lt.mc.Gravity) / (1 - lt.mc.Drag)
				x, y, z := vel.Mul(lt.conf.BlockCollisionVelocityMultiplier).Elem()
				// Calculate multipliers for all coordinates: 1 for the ones that
				// weren't on the same axis as the one collided with, -1 for the one
				// that was on that axis to deflect the projectile.
				mx, my, mz := hit.Face().Axis().Vec3().Mul(-2).Add(mgl64.Vec3{1, 1, 1}).Elem()

				vel = mgl64.Vec3{x * mx, y * my, z * mz}
			} else {
				vel = zeroVec3
			}
			end = hit.Position()
		}
	}
	return &Movement{v: viewers, e: e, pos: end, vel: vel, dpos: end.Sub(pos), dvel: vel.Sub(velBefore), rot: rot}, hit
}

// ignores returns a function to ignore entities in trace.Perform that are
// either a spectator, not living, the entity itself or its owner in the first
// 5 ticks.
func (lt *ProjectileBehaviour) ignores(e *Ent) trace.EntityFilter {
	return func(seq iter.Seq[world.Entity]) iter.Seq[world.Entity] {
		return func(yield func(world.Entity) bool) {
			for other := range seq {
				g, ok := other.(interface{ GameMode() world.GameMode })
				_, living := other.(Living)
				if (ok && !g.GameMode().HasCollision()) || e.H() == other.H() || !living || (e.data.Age < time.Second/4 && lt.conf.Owner == other.H()) {
					continue
				}
				if !yield(other) {
					return
				}
			}
		}
	}
}
