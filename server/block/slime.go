package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Slime is a storage block equivalent to nine slimeballs. It has both sticky and bouncy properties making it useful in
// conjunction with pistons to move both blocks and entities.
type Slime struct {
	solid
	transparent
}

// BreakInfo ...
func (s Slime) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(s))
}

// Friction ...
func (Slime) Friction() float64 {
	return 0.8
}

// EncodeItem ...
func (Slime) EncodeItem() (name string, meta int16) {
	return "minecraft:slime", 0
}

// EncodeBlock ...
func (Slime) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:slime", nil
}

// EntityLand ...
func (Slime) EntityLand(_ cube.Pos, _ *world.World, e world.Entity, distance *float64) {
	if s, ok := e.(interface {
		Sneaking() bool
	}); !ok || !s.Sneaking() {
		*distance = 0
	}
	if v, ok := e.(interface {
		Velocity() mgl64.Vec3
		SetVelocity(mgl64.Vec3)
	}); ok {
		vel := v.Velocity()
		vel[1] = -vel[1]
		v.SetVelocity(vel)
	}
}
