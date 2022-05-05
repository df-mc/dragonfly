package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EntityResult is the result of a ray trace collision with an entities bounding box.
type EntityResult struct {
	bb   cube.BBox
	pos  mgl64.Vec3
	face cube.Face

	entity world.Entity
}

// BBox returns the entities bounding box that was collided with.
func (r EntityResult) BBox() cube.BBox {
	return r.bb
}

// Position ...
func (r EntityResult) Position() mgl64.Vec3 {
	return r.pos
}

// Face ...
func (r EntityResult) Face() cube.Face {
	return r.face
}

// Entity returns the entity that was collided with.
func (r EntityResult) Entity() world.Entity {
	return r.entity
}

// EntityIntercept performs a ray trace and calculates the point on the entities bounding box's edge nearest to the start position
// that the ray collided with.
// EntityIntercept returns an EntityResult with the entity collided with and with the colliding vector closest to the start position,
// if no colliding point was found, a zero BlockResult is returned ok is false.
func EntityIntercept(e world.Entity, start, end mgl64.Vec3) (result EntityResult, ok bool) {
	bb := e.BBox().Translate(e.Position()).Grow(0.3)

	r, ok := BBoxIntercept(bb, start, end)
	if !ok {
		return
	}

	return EntityResult{bb: bb, pos: r.Position(), face: r.Face(), entity: e}, true
}
