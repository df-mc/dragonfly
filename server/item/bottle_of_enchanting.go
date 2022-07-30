package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// BottleOfEnchanting is a bottle that releases experience orbs when thrown.
type BottleOfEnchanting struct{}

// Use ...
func (b BottleOfEnchanting) Use(w *world.World, user User, ctx *UseContext) bool {
	splash, ok := world.EntityByName("minecraft:xp_bottle")
	if !ok {
		return false
	}

	p, ok := splash.(interface {
		New(pos, vel mgl64.Vec3, yaw, pitch float64) world.Entity
	})
	if !ok {
		return false
	}

	yaw, pitch := user.Rotation()
	e := p.New(eyePosition(user), directionVector(user).Mul(0.7), yaw, pitch)
	if o, ok := e.(owned); ok {
		o.Own(user)
	}

	ctx.SubtractFromCount(1)

	w.PlaySound(user.Position(), sound.ItemThrow{})
	w.AddEntity(e)
	return true
}

// EncodeItem ...
func (b BottleOfEnchanting) EncodeItem() (name string, meta int16) {
	return "minecraft:experience_bottle", 0
}
