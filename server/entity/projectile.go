package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"math/rand"
	"time"
)

type ProjectileBehaviourConfig struct {
	Gravity float64

	Drag float64

	// Damage specifies the damage dealt by the Projectile. If set to a negative
	// number, entities hitEntity are not hurt at all and are not knocked back.
	Damage float64

	KnockBackAddend float64

	Particle world.Particle

	ParticleCount int

	Sound world.Sound

	Critical bool // TODO: Change this to have a method that specifies if the arrow is critical and make that return false if collided.

	Hit func(e *Ent, target trace.Result)

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

	DisablePickup bool

	PickupItem item.Stack
}

func (conf ProjectileBehaviourConfig) New(owner world.Entity) *ProjectileBehaviour {
	if conf.ParticleCount == 0 {
		conf.ParticleCount = 1
	}
	return &ProjectileBehaviour{conf: conf, owner: owner, mc: &MovementComputer{
		Gravity:           conf.Gravity,
		Drag:              conf.Drag,
		DragBeforeGravity: true,
	}}
}

type ProjectileBehaviour struct {
	conf             ProjectileBehaviourConfig
	owner            world.Entity
	mc               *MovementComputer
	age, ageCollided int
	close            bool

	collisionPos cube.Pos
	collided     bool
}

func (lt *ProjectileBehaviour) Explode(e *Ent, src mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	e.vel = e.vel.Add(e.pos.Sub(src).Normalize().Mul(impact))
}

func (lt *ProjectileBehaviour) Tick(e *Ent) *Movement {
	if lt.close {
		_ = e.Close()
		return nil
	}
	w := e.World()

	e.mu.Lock()
	if lt.collided && lt.tickAttached(e) {
		e.mu.Unlock()
		return nil
	}
	before, vel := e.pos, e.vel
	m, result := lt.tickMovement(e)
	e.pos, e.vel = m.pos, m.vel

	lt.collisionPos, lt.collided = cube.Pos{}, false
	e.mu.Unlock()

	if result == nil {
		return m
	}

	if lt.conf.Particle != nil {
		for i := 0; i < lt.conf.ParticleCount; i++ {
			w.AddParticle(result.Position(), lt.conf.Particle)
		}
	}
	if lt.conf.Sound != nil {
		w.PlaySound(result.Position(), lt.conf.Sound)
	}

	switch r := result.(type) {
	case trace.EntityResult:
		if l, ok := r.Entity().(Living); ok && lt.conf.Damage >= 0 {
			lt.hitEntity(l, e, before, vel)
		}
	case trace.BlockResult:
		bpos := r.BlockPosition()
		if t, ok := w.Block(bpos).(block.TNT); ok && e.OnFireDuration() > 0 {
			t.Ignite(bpos, w)
		}
		if lt.conf.SurviveBlockCollision {
			lt.hitBlockSurviving(e, r, m)
			return m
		}
	}
	if lt.conf.Hit != nil {
		lt.conf.Hit(e, result)
	}

	lt.close = true
	return m
}

func (lt *ProjectileBehaviour) tickAttached(e *Ent) bool {
	w := e.World()
	boxes := w.Block(lt.collisionPos).Model().BBox(lt.collisionPos, w)
	box := e.Type().BBox(e).Translate(e.pos)

	for _, bb := range boxes {
		if box.IntersectsWith(bb.Translate(lt.collisionPos.Vec3()).Grow(0.05)) {
			if lt.ageCollided > 5 && !lt.conf.DisablePickup {
				lt.tryPickup(e)
			}
			lt.ageCollided++
			return true
		}
	}
	return false
}

// tryPickup checks for nearby projectile collectors and closes the entity if
// one was found.
func (lt *ProjectileBehaviour) tryPickup(e *Ent) {
	w := e.World()
	translated := e.Type().BBox(e).Translate(e.pos)
	grown := translated.GrowVec3(mgl64.Vec3{1, 0.5, 1})
	ignore := func(other world.Entity) bool {
		return e == other
	}
	for _, other := range w.EntitiesWithin(translated.Grow(2), ignore) {
		if !other.Type().BBox(other).Translate(other.Position()).IntersectsWith(grown) {
			continue
		}
		collector, ok := other.(Collector)
		if !ok {
			continue
		}
		// A collector was within range to pick up the entity.
		lt.close = true
		for _, viewer := range w.Viewers(e.pos) {
			viewer.ViewEntityAction(e, PickedUpAction{Collector: collector})
		}
		if lt.conf.PickupItem.Empty() {
			return
		}
		_ = collector.Collect(lt.conf.PickupItem)
	}
}

func (lt *ProjectileBehaviour) hitBlockSurviving(e *Ent, r trace.BlockResult, m *Movement) {
	x, y, z := m.vel.Mul(lt.conf.BlockCollisionVelocityMultiplier).Elem()
	// Calculate multipliers for all coordinates: 1 for the ones that
	// weren't on the same axis as the one collided with, -1 for the one
	// that was on that axis to deflect the projectile.
	mx, my, mz := r.Face().Axis().Vec3().Mul(-2).Add(mgl64.Vec3{1, 1, 1}).Elem()
	res := mgl64.Vec3{x * mx, y * my, z * mz}

	e.mu.Lock()
	e.vel = res
	e.mu.Unlock()

	if mgl64.FloatEqual(res.Len(), 0) {
		lt.collisionPos, lt.collided = r.BlockPosition(), true

		for _, v := range e.World().Viewers(m.pos) {
			v.ViewEntityAction(e, ArrowShakeAction{Duration: time.Millisecond * 350})
		}
	}
}

func (lt *ProjectileBehaviour) hitEntity(l Living, e *Ent, origin, vel mgl64.Vec3) {
	src := ProjectileDamageSource{Projectile: e, Owner: lt.owner}
	dmg := lt.conf.Damage * vel.Len()
	if lt.conf.Critical {
		dmg += rand.Float64() * dmg / 2
	}
	if _, vulnerable := l.Hurt(lt.conf.Damage, src); vulnerable {
		l.KnockBack(origin, 0.45+lt.conf.KnockBackAddend, 0.3608)

		if flammable, ok := l.(Flammable); ok && e.OnFireDuration() > 0 {
			flammable.SetOnFire(time.Second * 5)
		}
	}
}

func (lt *ProjectileBehaviour) tickMovement(e *Ent) (*Movement, trace.Result) {
	w, pos, vel, rot := e.World(), e.Position(), e.Velocity(), e.Rotation()
	viewers := w.Viewers(pos)

	velBefore := vel
	vel = lt.mc.applyHorizontalForces(w, pos, lt.mc.applyVerticalForces(vel))
	rot = cube.Rotation{
		mgl64.RadToDeg(math.Atan2(vel[0], vel[2])),
		mgl64.RadToDeg(math.Atan2(vel[1], math.Hypot(vel[0], vel[2]))),
	}

	var (
		end = pos.Add(vel)
		hit trace.Result
		ok  bool
	)
	if !mgl64.FloatEqual(end.Sub(pos).LenSqr(), 0) {
		if hit, ok = trace.Perform(pos, end, w, e.Type().BBox(e).Grow(1.0), lt.ignores(e)); ok {
			vel = zeroVec3
			end = hit.Position()
		}
	}
	lt.age++

	return &Movement{v: viewers, e: e, pos: end, vel: vel, dpos: end.Sub(pos), dvel: vel.Sub(velBefore), rot: rot}, hit
}

// ignores returns a function to ignore entities in trace.Perform that are
// either a spectator, not living, the entity itself or its owner in the first
// 5 ticks.
func (lt *ProjectileBehaviour) ignores(e *Ent) func(other world.Entity) bool {
	return func(other world.Entity) bool {
		g, ok := other.(interface{ GameMode() world.GameMode })
		_, living := other.(Living)
		return (ok && !g.GameMode().HasCollision()) || e == other || !living || (lt.age < 5 && lt.owner == other)
	}
}

// Projectile is a world.Entity that can be fired by another entity. It has an
// Owner method.
type Projectile interface {
	world.Entity
	Owner() world.Entity
}

// ProjectileComputer is used to compute movement of a projectile. When constructed, a MovementComputer must be passed.
type ProjectileComputer struct {
	*MovementComputer
	age int
}

// newProjectileComputer creates a ProjectileComputer with a gravity and drag
// value and if drag should be applied before gravity.
func newProjectileComputer(gravity, drag float64) *ProjectileComputer {
	return &ProjectileComputer{MovementComputer: &MovementComputer{
		Gravity:           gravity,
		Drag:              drag,
		DragBeforeGravity: true,
	}}
}

// TickMovement performs a movement tick on a projectile. Velocity is applied and changed according to the values
// of its Drag and Gravity. A ray trace is performed to see if the projectile has collided with any block or entity,
// the ray trace result is returned.
// The resulting Movement can be sent to viewers by calling Movement.Send.
func (c *ProjectileComputer) TickMovement(e Projectile, pos, vel mgl64.Vec3, yaw, pitch float64) (*Movement, trace.Result) {
	w := e.World()
	viewers := w.Viewers(pos)

	velBefore := vel
	vel = c.applyHorizontalForces(w, pos, c.applyVerticalForces(vel))
	end := pos.Add(vel)
	var hit trace.Result
	var ok bool
	if !mgl64.FloatEqual(end.Sub(pos).LenSqr(), 0) {
		hit, ok = trace.Perform(pos, end, w, e.Type().BBox(e).Grow(1.0), func(ent world.Entity) bool {
			g, ok := ent.(interface{ GameMode() world.GameMode })
			_, living := ent.(Living)
			return (ok && !g.GameMode().HasCollision()) || e == ent || (c.age < 5 && e.Owner() == ent) || !living
		})
	}
	if ok {
		vel = zeroVec3
		end = hit.Position()
	} else {
		yaw, pitch = mgl64.RadToDeg(math.Atan2(vel[0], vel[2])), mgl64.RadToDeg(math.Atan2(vel[1], math.Sqrt(vel[0]*vel[0]+vel[2]*vel[2])))
	}
	c.onGround = ok
	c.age++

	return &Movement{v: viewers, e: e,
		pos: end, vel: vel, dpos: end.Sub(pos), dvel: vel.Sub(velBefore),
		rot: cube.Rotation{yaw, pitch}, onGround: c.onGround,
	}, hit
}
