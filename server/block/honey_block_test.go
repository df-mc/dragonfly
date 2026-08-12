package block_test

import (
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// fakeFallEntity is a minimal world.Entity that tracks a fall distance and velocity, used to test
// block.HoneyBlock's fall damage reduction without spinning up a full player/entity.
type fakeFallEntity struct {
	fall float64
	vel  mgl64.Vec3
}

func (*fakeFallEntity) Close() error                 { return nil }
func (*fakeFallEntity) H() *world.EntityHandle       { return nil }
func (*fakeFallEntity) Position() mgl64.Vec3         { return mgl64.Vec3{} }
func (*fakeFallEntity) Rotation() cube.Rotation      { return cube.Rotation{} }
func (e *fakeFallEntity) FallDistance() float64      { return e.fall }
func (e *fakeFallEntity) ResetFallDistance()         { e.fall = 0 }
func (e *fakeFallEntity) Velocity() mgl64.Vec3       { return e.vel }
func (e *fakeFallEntity) SetVelocity(vel mgl64.Vec3) { e.vel = vel }

// TestHoneyBlockEntityLandReducesNotAbsorbsFallDistance verifies that landing on a honey block
// only reduces the effective fall distance used for damage by 80%, rather than absorbing it
// entirely (which would negate all fall damage regardless of fall height).
func TestHoneyBlockEntityLandReducesNotAbsorbsFallDistance(t *testing.T) {
	e := &fakeFallEntity{fall: 10}
	distance := e.fall

	block.HoneyBlock{}.EntityLand(cube.Pos{}, nil, e, &distance)

	// Without honey, damage-relevant distance would be 10-3=7. With an 80% reduction, the
	// damage-relevant distance (distance-3) should be 0.2*7=1.4, i.e. distance should be 4.4.
	if want := 4.4; !mgl64.FloatEqualThreshold(distance, want, 1e-9) {
		t.Errorf("expected reduced distance %v, got %v", want, distance)
	}
	if distance == 0 {
		t.Errorf("honey block must not absorb all fall distance")
	}
}

// TestHoneyBlockEntityInsideSlowsSlideAndResetsFallDistance verifies that an entity sliding down the
// side of a honey block has its downward velocity capped and its fall distance reset, so it descends
// slowly and takes no fall damage.
func TestHoneyBlockEntityInsideSlowsSlideAndResetsFallDistance(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	w.Do(func(tx *world.Tx) {
		e := &fakeFallEntity{fall: 10, vel: mgl64.Vec3{0.4, -0.5, 0.4}}

		block.HoneyBlock{}.EntityInside(cube.Pos{}, tx, e)

		if want := -0.05; !mgl64.FloatEqualThreshold(e.Velocity()[1], want, 1e-9) {
			t.Errorf("expected downward velocity to be capped at %v, got %v", want, e.Velocity()[1])
		}
		// Falling faster than -0.13 also scales horizontal momentum by -0.05/velY = 0.1.
		if want := 0.04; !mgl64.FloatEqualThreshold(e.Velocity()[0], want, 1e-9) {
			t.Errorf("expected horizontal velocity to be scaled to %v, got %v", want, e.Velocity()[0])
		}
		if e.fall != 0 {
			t.Errorf("expected fall distance to be reset to 0, got %v", e.fall)
		}
	})
}

// TestHoneyBlockEntityInsideIgnoresUpwardMovement verifies that an entity moving upwards fast enough
// (for example, jumping past the block) is not caught and slowed by the sliding logic.
func TestHoneyBlockEntityInsideIgnoresUpwardMovement(t *testing.T) {
	w := world.Config{Synchronous: true, Entities: entity.DefaultRegistry}.New()
	defer w.Close()

	w.Do(func(tx *world.Tx) {
		vel := mgl64.Vec3{0, 0.42, 0}
		e := &fakeFallEntity{fall: 10, vel: vel}

		block.HoneyBlock{}.EntityInside(cube.Pos{}, tx, e)

		if e.Velocity() != vel {
			t.Errorf("expected upward velocity %v to be left untouched, got %v", vel, e.Velocity())
		}
	})
}
