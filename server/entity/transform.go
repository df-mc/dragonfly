package entity

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"sync"
)

// transform holds the base position and velocity of an entity. It holds several methods which can be used when
// embedding the struct.
type transform struct {
	e        world.Entity
	mu       sync.Mutex
	vel, pos mgl64.Vec3
}

// newTransform creates a new transform to embed for the world.Entity passed.
func newTransform(e world.Entity, pos mgl64.Vec3) transform {
	return transform{e: e, pos: pos}
}

// Position returns the current position of the entity.
func (t *transform) Position() mgl64.Vec3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.pos
}

// Velocity returns the current velocity of the entity. The values in the Vec3 returned represent the speed on
// that axis in blocks/tick.
func (t *transform) Velocity() mgl64.Vec3 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.vel
}

// SetVelocity sets the velocity of the entity. The values in the Vec3 passed represent the speed on
// that axis in blocks/tick.
func (t *transform) SetVelocity(v mgl64.Vec3) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.vel = v
}

// Rotation always returns 0.
func (t *transform) Rotation() (float64, float64) { return 0, 0 }

// World returns the world of the entity.
func (t *transform) World() *world.World {
	w, _ := world.OfEntity(t.e)
	return w
}

// Close closes the transform and removes the associated entity from the world.
func (t *transform) Close() error {
	w, _ := world.OfEntity(t.e)
	w.RemoveEntity(t.e)
	return nil
}
