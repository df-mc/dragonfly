package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

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
		yaw: yaw, pitch: pitch, onGround: c.onGround,
	}, hit
}
