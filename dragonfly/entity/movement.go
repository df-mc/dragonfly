package entity

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// boxes returns the axis aligned bounding box of a block.
func boxes(b world.Block, pos world.BlockPos, w *world.World) []physics.AABB {
	return b.Model().AABB(pos, w)
}

// movementComputer is used to compute movement of an entity. When constructed, the gravity of the entity
// the movement is computed for must be passed.
type movementComputer struct {
	onGround          bool
	gravity           float64
	dragBeforeGravity bool
}

// tickMovement performs a movement tick on an entity. Velocity is applied and changed according to the values
// of its drag and gravity.
// The new position of the entity after movement is returned.
func (c *movementComputer) tickMovement(e world.Entity) mgl64.Vec3 {
	viewers := e.World().Viewers(e.Position())
	if !e.Velocity().ApproxEqualThreshold(mgl64.Vec3{}, 0.001) {
		for _, v := range viewers {
			v.ViewEntityVelocity(e, e.Velocity())
		}
	}

	toMove, velocity := c.handleCollision(e)
	e.SetVelocity(velocity)
	v := c.move(e, toMove, viewers)
	e.SetVelocity(c.applyGravity(e))
	e.SetVelocity(c.applyFriction(e))
	return v
}

// applyGravity applies gravity to the entity's velocity. By default, 0.08 is subtracted from the y value, or
// a different value if the Gravity
func (c *movementComputer) applyGravity(e world.Entity) mgl64.Vec3 {
	velocity := e.Velocity()
	if c.dragBeforeGravity {
		velocity[1] *= 0.98
	}
	velocity[1] -= c.gravity
	if !c.dragBeforeGravity {
		velocity[1] *= 0.98
	}
	return velocity
}

// applyFriction applies friction to the entity, reducing its velocity on the X and Z axes.
func (c *movementComputer) applyFriction(e world.Entity) mgl64.Vec3 {
	velocity := e.Velocity()
	if c.onGround {
		velocity[0] *= 0.6
		velocity[2] *= 0.6
		return velocity
	}
	velocity[0] *= 0.91
	velocity[2] *= 0.91
	return velocity
}

// move moves the entity so that all viewers in the world can see it, adding the velocity to the position.
func (c *movementComputer) move(e world.Entity, deltaPos mgl64.Vec3, viewers []world.Viewer) mgl64.Vec3 {
	if deltaPos.ApproxEqualThreshold(mgl64.Vec3{}, 0.01) {
		return e.Position()
	}
	for _, v := range viewers {
		v.ViewEntityMovement(e, deltaPos, 0, 0)
	}
	return e.Position().Add(deltaPos)
}

// handleCollision handles the collision of the entity with blocks, adapting the velocity of the entity if it
// happens to collide with a block.
// The final velocity and the Vec3 that the entity should move is returned.
func (c *movementComputer) handleCollision(e world.Entity) (move mgl64.Vec3, velocity mgl64.Vec3) {
	// TODO: Implement collision with other entities.
	velocity = e.Velocity()
	deltaX, deltaY, deltaZ := velocity[0], velocity[1], velocity[2]

	// Entities only ever have a single bounding box.
	entityAABB := e.AABB().Translate(e.Position())
	blocks := blockAABBsAround(e, entityAABB.Extend(velocity))

	if !mgl64.FloatEqual(deltaY, 0) {
		// First we move the entity AABB on the Y axis.
		for _, blockAABB := range blocks {
			deltaY = entityAABB.CalculateYOffset(blockAABB, deltaY)
		}
		entityAABB = entityAABB.Translate(mgl64.Vec3{0, deltaY})
	}
	if !mgl64.FloatEqual(deltaX, 0) {
		// Then on the X axis.
		for _, blockAABB := range blocks {
			deltaX = entityAABB.CalculateXOffset(blockAABB, deltaX)
		}
		entityAABB = entityAABB.Translate(mgl64.Vec3{deltaX})
	}
	if !mgl64.FloatEqual(deltaZ, 0) {
		// And finally on the Z axis.
		for _, blockAABB := range blocks {
			deltaZ = entityAABB.CalculateZOffset(blockAABB, deltaZ)
		}
	}
	if !mgl64.FloatEqual(velocity[0], 0) {
		// The Y velocity of the entity is currently not 0, meaning it is moving either up or down. We can
		// then assume the entity is not currently on the ground.
		c.onGround = false
	}
	if !mgl64.FloatEqual(deltaX, velocity[0]) {
		velocity[0] = 0
	}
	if !mgl64.FloatEqual(deltaY, velocity[1]) {
		// The entity either hit the ground or hit the ceiling.
		if velocity[1] < 0 {
			// The entity was going down, so we can assume it is now on the ground.
			c.onGround = true
		}
		velocity[1] = 0
	}
	if !mgl64.FloatEqual(deltaZ, velocity[2]) {
		velocity[2] = 0
	}
	return mgl64.Vec3{deltaX, deltaY, deltaZ}, velocity
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
				pos := world.BlockPos{x, y, z}
				boxes := boxes(e.World().Block(pos), pos, w)
				for _, box := range boxes {
					blockAABBs = append(blockAABBs, box.Translate(mgl64.Vec3{float64(x), float64(y), float64(z)}))
				}
			}
		}
	}
	return blockAABBs
}

// OnGround checks if the entity that this computer calculates is currently on the ground.
func (c *movementComputer) OnGround() bool {
	return c.onGround
}
