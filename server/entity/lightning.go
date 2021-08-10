package entity

import (
	"github.com/df-mc/dragonfly/server/entity/physics"
	math "github.com/df-mc/dragonfly/server/internal"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"sync/atomic"
)

// Lightning is a lethal element to thunderstorms. Lightning momentarily increases the skylight's brightness to slightly greater than full daylight.
type Lightning struct {
	pos atomic.Value

	state    int
	liveTime int
}

// NewLightning creates a lightning entity. The lightning entity will be positioned at the position passed.
func NewLightning(pos mgl64.Vec3) *Lightning {
	li := &Lightning{
		state:    2,
		liveTime: math.NextIntn(3) + 1,
	}
	li.pos.Store(pos)

	// TODO set blocks on fire

	return li
}

// Position returns the current position of the lightning entity.
func (li *Lightning) Position() mgl64.Vec3 {
	return li.pos.Load().(mgl64.Vec3)
}

// World returns the world that the lightning entity is currently in, or nil if it is not added to a world.
func (li *Lightning) World() *world.World {
	w, _ := world.OfEntity(li)
	return w
}

// AABB ...
func (Lightning) AABB() physics.AABB {
	return physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{})
}

// State ...
func (li *Lightning) State() int {
	return li.state
}

// Close closes the lighting.
func (li *Lightning) Close() error {
	if li.World() != nil {
		li.World().RemoveEntity(li)
	}

	return nil
}

// OnGround ...
func (Lightning) OnGround() bool {
	return false
}

// Rotation ...
func (li *Lightning) Rotation() (yaw, pitch float64) {
	return 0, 0
}

// EncodeEntity ...
func (li *Lightning) EncodeEntity() string {
	return "minecraft:lightning_bolt"
}

// Name ...
func (li *Lightning) Name() string {
	return "Lightning Bolt"
}

// Tick ...
func (li *Lightning) Tick(_ int64) {
	if li.state == 2 {
		li.World().PlaySound(li.Position(), sound.Thunder{})
		li.World().PlaySound(li.Position(), sound.Explosion{})
	}

	li.state--

	if li.state < 0 {
		if li.liveTime == 0 {
			_ = li.Close()
			return
		} else if li.state < -math.NextIntn(10) {
			li.liveTime--
			li.state = 1

			// TODO set blocks on fire, again?
		}
	}

	if li.state >= 0 {
		bb := li.AABB().Grow(3)
		bb.Extend(mgl64.Vec3{bb.Max().X() + 6})

		// TODO get all colliding entities in the aabb, and then damage them
	}
}
