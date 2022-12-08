package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

type ProjectileLifetimeConfig struct {
	Owner world.Entity

	Gravity float64

	Drag float64

	Damage float64

	Particle world.Particle
}

func (conf ProjectileLifetimeConfig) New() *ProjectileLifetime {
	return &ProjectileLifetime{conf: conf, mc: &MovementComputer{
		Gravity:           conf.Gravity,
		Drag:              conf.Drag,
		DragBeforeGravity: true,
	}}
}

type ProjectileLifetime struct {
	conf  ProjectileLifetimeConfig
	mc    *MovementComputer
	age   int
	close bool
}

func (lt *ProjectileLifetime) Explode(e *Ent, src mgl64.Vec3, impact float64, _ block.ExplosionConfig) {
	e.vel = e.vel.Add(e.pos.Sub(src).Normalize().Mul(impact))
}

func (lt *ProjectileLifetime) Tick(e *Ent) *Movement {
	if lt.close {
		_ = e.Close()
		return nil
	}
	m, result := lt.tickMovement(e)
	if result == nil {
		return m
	}

	if lt.conf.Particle != nil {
		for i := 0; i < 6; i++ {
			e.World().AddParticle(result.Position(), lt.conf.Particle)
		}
	}

	if r, ok := result.(trace.EntityResult); ok {
		if l, ok := r.Entity().(Living); ok {
			src := ProjectileDamageSource{Projectile: e, Owner: lt.conf.Owner}
			if _, vulnerable := l.Hurt(lt.conf.Damage, src); vulnerable {
				l.KnockBack(m.pos, 0.45, 0.3608)
			}
		}
	}

	lt.close = true
	return m
}

func (lt *ProjectileLifetime) tickMovement(e *Ent) (*Movement, trace.Result) {
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
func (lt *ProjectileLifetime) ignores(e *Ent) func(other world.Entity) bool {
	return func(other world.Entity) bool {
		g, ok := other.(interface{ GameMode() world.GameMode })
		_, living := other.(Living)
		return (ok && !g.GameMode().HasCollision()) || e == other || !living || (lt.age < 5 && lt.conf.Owner == other)
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
