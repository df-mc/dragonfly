package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
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

	p, ok := splash.(projectile)
	if !ok {
		return false
	}

	yaw, pitch := user.Rotation()
	e := p.New(eyePosition(user), directionVector(user).Mul(0.5), yaw, pitch)
	if o, ok := e.(owned); ok {
		o.Own(user)
	}
	if pot, ok := e.(splashPotion); ok {
		pot.SetVariant(s.Type)
	}

	ctx.SubtractFromCount(1)

	w.PlaySound(user.Position(), sound.ItemThrow{})

	w.AddEntity(e)

	return true
}

// EncodeItem ...
func (s SplashPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:splash_potion", int16(s.Type.Uint8())
}

// splashPotion represents an entity instance of a SplashPotion.
type splashPotion interface {
	// SetVariant sets the variant of the splash potion.
	SetVariant(variant potion.Potion)
	// Variant returns the variant of the splash potion.
	Variant() potion.Potion
}
