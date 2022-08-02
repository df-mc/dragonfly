package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// SplashPotion is an item that grants effects when thrown.
type SplashPotion struct {
	// Type is the type of splash potion.
	Type potion.Potion
}

// MaxCount ...
func (s SplashPotion) MaxCount() int {
	return 1
}

// Use ...
func (s SplashPotion) Use(w *world.World, user User, ctx *UseContext) bool {
	splash, ok := world.EntityByName("minecraft:splash_potion")
	if !ok {
		return false
	}

	p, ok := splash.(interface {
		New(pos, vel mgl64.Vec3, t potion.Potion, owner world.Entity) world.Entity
	})
	if !ok {
		return false
	}

	w.PlaySound(user.Position(), sound.ItemThrow{})
	w.AddEntity(p.New(eyePosition(user), directionVector(user).Mul(0.5), s.Type, user))

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (s SplashPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:splash_potion", int16(s.Type.Uint8())
}
