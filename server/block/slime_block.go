package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// SlimeBlock is a block that slows down entities and bounces them up if they drop onto it.
type SlimeBlock struct {
	solid
	transparent
}

// velocityEntity is an entity that has velocity.
type velocityEntity interface {
	// Velocity returns the current velocity of the entity. The values in the Vec3 returned represent the speed on
	// that axis in blocks/tick.
	Velocity() mgl64.Vec3
	// SetVelocity sets the velocity of the entity. The values in the Vec3 passed represent the speed on
	// that axis in blocks/tick.
	SetVelocity(v mgl64.Vec3)
}

// EntityLand ...
func (SlimeBlock) EntityLand(_ cube.Pos, _ *world.World, e world.Entity) {
	if e, ok := e.(fallDistanceEntity); ok {
		e.ResetFallDistance()
	}
	if e, ok := e.(velocityEntity); ok {
		v := e.Velocity()
		v[1] *= -1
		e.SetVelocity(v)
	}
}

// BreakInfo ...
func (s SlimeBlock) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(s))
}

// Friction ...
func (SlimeBlock) Friction() float64 {
	return 0.8
}

// EncodeItem ...
func (SlimeBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:slime", 0
}

// EncodeBlock ...
func (SlimeBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:slime", nil
}
