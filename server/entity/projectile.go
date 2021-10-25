package entity

import (
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// Owned represents an entity that is "owned" by another entity. Entities like projectiles typically are "owned".
type Owned interface {
	world.Entity
	Owner() world.Entity
	Own(owner world.Entity)
}

// Projectile represents an entity that can be launched.
type Projectile interface {
	world.Entity
	// New creates a new projectile with the position, velocity, yaw, and pitch provided. It does not spawn
	// the projectile.
	New(pos, vel mgl64.Vec3, yaw, pitch float64) world.Entity
}

// ProjectileComputer is used to compute movement of a projectile. When constructed, a MovementComputer must be passed.
type ProjectileComputer struct {
	*MovementComputer
}

// TickMovement performs a movement tick on a projectile. Velocity is applied and changed according to the values
// of its Drag and Gravity. A ray trace is performed to see if the projectile has collided with any block or entity,
// the ray trace result is returned.
func (c *ProjectileComputer) TickMovement(e Projectile, pos, vel mgl64.Vec3, yaw, pitch float64, ignored func(world.Entity) bool) (mgl64.Vec3, mgl64.Vec3, float64, float64, trace.Result) {
	w := e.World()
	viewers := w.Viewers(pos)

	vel = c.applyHorizontalForces(w, pos, c.applyVerticalForces(vel))
	end := pos.Add(vel)
	hit, ok := trace.Perform(pos, end, w, e.AABB().Grow(1.0), ignored)
	if ok {
		vel = zeroVec3
		end = hit.Position()
	} else {
		yaw, pitch = math.Atan2(vel[0], vel[2])*180/math.Pi, math.Atan2(vel[1], math.Sqrt(vel[0]*vel[0]+vel[2]*vel[2]))*180/math.Pi
	}
	c.onGround = ok

	c.sendMovement(e, viewers, end, end.Sub(pos), vel, yaw, pitch)

	return end, vel, yaw, pitch, hit
}
