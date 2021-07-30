package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// MovementComputer is used to compute movement of an entity. When constructed, the Gravity of the entity
// the movement is computed for must be passed.
type MovementComputer struct {
	Gravity           float64
	DragBeforeGravity bool
	Drag              float64

	onGround bool
}

// TickMovement performs a movement tick on an entity. Velocity is applied and changed according to the values
// of its Drag and Gravity.
// The new position of the entity after movement is returned.
func (c *MovementComputer) TickMovement(e world.Entity, pos, vel mgl64.Vec3, yaw, pitch float64) (mgl64.Vec3, mgl64.Vec3) {
	viewers := e.World().Viewers(pos)

	vel = c.applyHorizontalForces(c.applyVerticalForces(vel))
	dPos, vel := c.checkCollision(e, pos, vel)

	c.sendMovement(e, viewers, pos, dPos, vel, yaw, pitch)

	if dPos.ApproxEqualThreshold(zeroVec3, epsilon) {
		return pos, vel
	}
	return pos.Add(dPos), vel
}

// OnGround checks if the entity that this computer calculates is currently on the ground.
func (c *MovementComputer) OnGround() bool {
	return c.onGround
}

// zeroVec3 is a mgl64.Vec3 with zero values.
var zeroVec3 mgl64.Vec3

// epsilon is the epsilon used for thresholds for change used for change in position and velocity.
var epsilon = 0.001

// sendMovement sends the movement of the world.Entity passed (dPos and vel) to all viewers passed.
func (c *MovementComputer) sendMovement(e world.Entity, viewers []world.Viewer, pos, dPos, vel mgl64.Vec3, yaw, pitch float64) {
	posChanged := !dPos.ApproxEqualThreshold(zeroVec3, epsilon)
	velChanged := !vel.ApproxEqualThreshold(zeroVec3, epsilon)
	for _, v := range viewers {
		// TODO: Don't always send velocity and position change. This causes very jittery movement client-side
		//  which looks awful.
		if velChanged {
			v.ViewEntityVelocity(e, vel)
		}
		if posChanged {
			v.ViewEntityMovement(e, pos, yaw, pitch, c.onGround)
		}
	}
}

// applyVerticalForces applies gravity and drag on the Y axis, based on the Gravity and Drag values set.
func (c *MovementComputer) applyVerticalForces(vel mgl64.Vec3) mgl64.Vec3 {
	if c.DragBeforeGravity {
		vel[1] *= 1 - c.Drag
	}
	vel[1] -= c.Gravity
	if !c.DragBeforeGravity {
		vel[1] *= 1 - c.Drag
	}
	return vel
}

// applyHorizontalForces applies friction to the velocity based on the Drag value, reducing it on the X and Z axes.
func (c *MovementComputer) applyHorizontalForces(vel mgl64.Vec3) mgl64.Vec3 {
	friction := 1 - c.Drag
	if c.onGround {
		friction = 0.6
	}
	vel[0] *= friction
	vel[2] *= friction
	return vel
}

// checkCollision handles the collision of the entity with blocks, adapting the velocity of the entity if it
// happens to collide with a block.
// The final velocity and the Vec3 that the entity should move is returned.
func (c *MovementComputer) checkCollision(e world.Entity, pos, vel mgl64.Vec3) (mgl64.Vec3, mgl64.Vec3) {
	// TODO: Implement collision with other entities.
	deltaX, deltaY, deltaZ := vel[0], vel[1], vel[2]

	// Entities only ever have a single bounding box.
	entityAABB := e.AABB().Translate(pos)
	blocks := blockAABBsAround(e, entityAABB.Extend(vel))

	if !mgl64.FloatEqualThreshold(deltaY, 0, epsilon) {
		// First we move the entity AABB on the Y axis.
		for _, blockAABB := range blocks {
			deltaY = entityAABB.CalculateYOffset(blockAABB, deltaY)
		}
		entityAABB = entityAABB.Translate(mgl64.Vec3{0, deltaY})
	}
	if !mgl64.FloatEqualThreshold(deltaX, 0, epsilon) {
		// Then on the X axis.
		for _, blockAABB := range blocks {
			deltaX = entityAABB.CalculateXOffset(blockAABB, deltaX)
		}
		entityAABB = entityAABB.Translate(mgl64.Vec3{deltaX})
	}
	if !mgl64.FloatEqualThreshold(deltaZ, 0, epsilon) {
		// And finally on the Z axis.
		for _, blockAABB := range blocks {
			deltaZ = entityAABB.CalculateZOffset(blockAABB, deltaZ)
		}
	}
	if !mgl64.FloatEqual(vel[1], 0) {
		// The Y velocity of the entity is currently not 0, meaning it is moving either up or down. We can
		// then assume the entity is not currently on the ground.
		c.onGround = false
	}
	if !mgl64.FloatEqual(deltaX, vel[0]) {
		vel[0] = 0
	}
	if !mgl64.FloatEqual(deltaY, vel[1]) {
		// The entity either hit the ground or hit the ceiling.
		if vel[1] < 0 {
			// The entity was going down, so we can assume it is now on the ground.
			c.onGround = true
		}
		vel[1] = 0
	}
	if !mgl64.FloatEqual(deltaZ, vel[2]) {
		vel[2] = 0
	}
	return mgl64.Vec3{deltaX, deltaY, deltaZ}, vel
}

// blockAABBsAround returns all blocks around the entity passed, using the AABB passed to make a prediction of
// what blocks need to have their AABB returned.
func blockAABBsAround(e world.Entity, aabb physics.AABB) []physics.AABB {
	w := e.World()
	grown := aabb.Grow(0.25)
	min, max := grown.Min(), grown.Max()
	minX, minY, minZ := int(math.Floor(min[0])), int(math.Floor(min[1])), int(math.Floor(min[2]))
	maxX, maxY, maxZ := int(math.Ceil(max[0])), int(math.Ceil(max[1])), int(math.Ceil(max[2]))

	// A prediction of one AABB per block, plus an additional 2, in case
	blockAABBs := make([]physics.AABB, 0, (maxX-minX)*(maxY-minY)*(maxZ-minZ)+2)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				pos := cube.Pos{x, y, z}
				boxes := boxes(w.Block(pos), pos, w)
				for _, box := range boxes {
					blockAABBs = append(blockAABBs, box.Translate(mgl64.Vec3{float64(x), float64(y), float64(z)}))
				}
			}
		}
	}
	return blockAABBs
}

// boxes returns the axis aligned bounding box of a block.
func boxes(b world.Block, pos cube.Pos, w *world.World) []physics.AABB {
	return b.Model().AABB(pos, w)
}
