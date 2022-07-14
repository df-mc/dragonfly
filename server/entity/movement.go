package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math"
)

// MovementComputer is used to compute movement of an entity. When constructed, the Gravity of the entity
// the movement is computed for must be passed.
type MovementComputer struct {
	Gravity, Drag     float64
	DragBeforeGravity bool

	onGround bool
}

// Movement represents the movement of a world.Entity as a result of a call to MovementComputer.TickMovement. The
// resulting position and velocity can be obtained by calling Position and Velocity. These can be sent to viewers by
// calling Send.
type Movement struct {
	v                    []world.Viewer
	e                    world.Entity
	pos, vel, dpos, dvel mgl64.Vec3
	yaw, pitch           float64
	onGround             bool
}

// Send sends the Movement to any viewers watching the entity at the time of the movement. If the position/velocity
// changes were negligible, nothing is sent.
func (m *Movement) Send() {
	posChanged := !m.dpos.ApproxEqualThreshold(zeroVec3, epsilon)
	velChanged := !m.dvel.ApproxEqualThreshold(zeroVec3, epsilon)

	for _, v := range m.v {
		if posChanged {
			v.ViewEntityMovement(m.e, m.pos, m.yaw, m.pitch, m.onGround)
		}
		if velChanged {
			v.ViewEntityVelocity(m.e, m.vel)
		}
	}
}

// Position returns the position as a result of the Movement as an mgl64.Vec3.
func (m *Movement) Position() mgl64.Vec3 {
	return m.pos
}

// Velocity returns the velocity after the Movement as an mgl64.Vec3.
func (m *Movement) Velocity() mgl64.Vec3 {
	return m.vel
}

// Rotation returns the rotation, yaw and pitch, of the entity after the Movement.
func (m *Movement) Rotation() (yaw, pitch float64) {
	return m.yaw, m.pitch
}

// TickMovement performs a movement tick on an entity. Velocity is applied and changed according to the values
// of its Drag and Gravity.
// The new position of the entity after movement is returned.
// The resulting Movement can be sent to viewers by calling Movement.Send.
func (c *MovementComputer) TickMovement(e world.Entity, pos, vel mgl64.Vec3, yaw, pitch float64) *Movement {
	w := e.World()
	viewers := w.Viewers(pos)

	velBefore := vel
	vel = c.applyHorizontalForces(w, pos, c.applyVerticalForces(vel))
	dPos, vel := c.checkCollision(e, pos, vel)

	return &Movement{v: viewers, e: e,
		pos: pos.Add(dPos), vel: vel, dpos: dPos, dvel: vel.Sub(velBefore),
		yaw: yaw, pitch: pitch, onGround: c.onGround,
	}
}

// OnGround checks if the entity that this computer calculates is currently on the ground.
func (c *MovementComputer) OnGround() bool {
	return c.onGround
}

// zeroVec3 is a mgl64.Vec3 with zero values.
var zeroVec3 mgl64.Vec3

// epsilon is the epsilon used for thresholds for change used for change in position and velocity.
const epsilon = 0.001

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
func (c *MovementComputer) applyHorizontalForces(w *world.World, pos, vel mgl64.Vec3) mgl64.Vec3 {
	friction := 1 - c.Drag
	if c.onGround {
		if f, ok := w.Block(cube.PosFromVec3(pos).Side(cube.FaceDown)).(interface {
			Friction() float64
		}); ok {
			friction *= f.Friction()
		} else {
			friction *= 0.6
		}
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
	entityBBox := e.BBox().Translate(pos)
	blocks := blockBBoxsAround(e, entityBBox.Extend(vel))

	if !mgl64.FloatEqualThreshold(deltaY, 0, epsilon) {
		// First we move the entity BBox on the Y axis.
		for _, blockBBox := range blocks {
			deltaY = entityBBox.YOffset(blockBBox, deltaY)
		}
		entityBBox = entityBBox.Translate(mgl64.Vec3{0, deltaY})
	}
	if !mgl64.FloatEqualThreshold(deltaX, 0, epsilon) {
		// Then on the X axis.
		for _, blockBBox := range blocks {
			deltaX = entityBBox.XOffset(blockBBox, deltaX)
		}
		entityBBox = entityBBox.Translate(mgl64.Vec3{deltaX})
	}
	if !mgl64.FloatEqualThreshold(deltaZ, 0, epsilon) {
		// And finally on the Z axis.
		for _, blockBBox := range blocks {
			deltaZ = entityBBox.ZOffset(blockBBox, deltaZ)
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

// blockBBoxsAround returns all blocks around the entity passed, using the BBox passed to make a prediction of
// what blocks need to have their BBox returned.
func blockBBoxsAround(e world.Entity, box cube.BBox) []cube.BBox {
	w := e.World()
	grown := box.Grow(0.25)
	min, max := grown.Min(), grown.Max()
	minX, minY, minZ := int(math.Floor(min[0])), int(math.Floor(min[1])), int(math.Floor(min[2]))
	maxX, maxY, maxZ := int(math.Ceil(max[0])), int(math.Ceil(max[1])), int(math.Ceil(max[2]))

	// A prediction of one BBox per block, plus an additional 2, in case
	blockBBoxs := make([]cube.BBox, 0, (maxX-minX)*(maxY-minY)*(maxZ-minZ)+2)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				pos := cube.Pos{x, y, z}
				boxes := w.Block(pos).Model().BBox(pos, w)
				for _, box := range boxes {
					blockBBoxs = append(blockBBoxs, box.Translate(mgl64.Vec3{float64(x), float64(y), float64(z)}))
				}
			}
		}
	}
	return blockBBoxs
}
