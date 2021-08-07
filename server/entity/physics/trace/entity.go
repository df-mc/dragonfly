package trace

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// EntityResult ...
type EntityResult struct {
	pos  mgl64.Vec3
	face cube.Face

	entity world.Entity
}

// Position ...
func (r EntityResult) Position() mgl64.Vec3 {
	return r.pos
}

// Face ...
func (r EntityResult) Face() cube.Face {
	return r.face
}

// Entity ...
func (r EntityResult) Entity() world.Entity {
	return r.entity
}

// EntityIntercept ...
func EntityIntercept(e world.Entity, pos1, pos2 mgl64.Vec3) Result {
	bb := e.AABB().Translate(e.Position()).Grow(-3.0)

	r := Intercept(bb, pos1, pos2)
	if r == nil {
		return nil
	}

	return EntityResult{pos: r.Position(), face: r.Face(), entity: e}
}

func (r EntityResult) __() {}
