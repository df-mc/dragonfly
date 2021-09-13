package entity

import (
	"github.com/df-mc/dragonfly/server/entity/physics/trace"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

type projectileComputer struct {
	c *MovementComputer
}

func (c *projectileComputer) TickMovement(e world.Entity, pos, vel mgl64.Vec3, yaw, pitch float64, ignoredEntities ...world.Entity) (mgl64.Vec3, mgl64.Vec3, float64, float64, trace.Result) {
	w := e.World()
	viewers := w.Viewers(pos)

	vel = c.c.applyHorizontalForces(c.c.applyVerticalForces(vel))
	end := pos.Add(vel)
	hit, ok := trace.Perform(pos, end, w, e.AABB().Grow(1.0), append(ignoredEntities, e)...)

	c.c.onGround = ok
	if ok {
		vel = zeroVec3
		end = hit.Position()
	} else {
		yaw, pitch = math.Atan2(vel[0], vel[2])*180/math.Pi, math.Atan2(vel[1], math.Sqrt(vel[0]*vel[0]+vel[2]*vel[2]))*180/math.Pi
	}

	c.c.sendMovement(e, viewers, end, end.Sub(pos), vel, yaw, pitch)

	return end, vel, yaw, pitch, hit
}

func (c *projectileComputer) OnGround() bool {
	return c.c.OnGround()
}
