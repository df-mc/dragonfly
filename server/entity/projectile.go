package entity

import (
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// ProjectileComputer is used to compute movement of a projectile. When constructed, a MovementComputer must be passed.
type ProjectileComputer struct {
	*MovementComputer
}

// TickMovement performs a movement tick on a projectile. Velocity is applied and changed according to the values
// of its Drag and Gravity. A ray trace is performed to see if the projectile has collided with any block or entity,
// the ray trace result is returned.
func (c *ProjectileComputer) TickMovement(e world.Entity, pos, vel mgl64.Vec3, yaw, pitch float64, ignoredEntities ...world.Entity) (mgl64.Vec3, mgl64.Vec3, float64, float64, trace.Result) {
	w := e.World()
	viewers := w.Viewers(pos)

	vel = c.applyHorizontalForces(w, pos, c.applyVerticalForces(vel))
	end := pos.Add(vel)
	hit, ok := trace.Perform(pos, end, w, e.AABB().Grow(1.0), append(ignoredEntities, e)...)

	c.onGround = ok
	if ok {
		vel = zeroVec3
		end = hit.Position()
	} else {
		yaw, pitch = math.Atan2(vel[0], vel[2])*180/math.Pi, math.Atan2(vel[1], math.Sqrt(vel[0]*vel[0]+vel[2]*vel[2]))*180/math.Pi
	}

	c.sendMovement(e, viewers, end, end.Sub(pos), vel, yaw, pitch)

	return end, vel, yaw, pitch, hit
}
