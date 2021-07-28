package entity

import (
	"github.com/go-gl/mathgl/mgl64"
	"sync"
)

// transform holds the base position and velocity of an entity. It holds several methods which can be used when
// embedding the struct.
type transform struct {
	mu       sync.Mutex
	vel, pos mgl64.Vec3
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
