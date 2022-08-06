package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Egg is an item that can be used to craft food items, or as a throwable entity to spawn chicks.
type Egg struct{}

// MaxCount ...
func (e Egg) MaxCount() int {
	return 16
}

// Use ...
func (e Egg) Use(w *world.World, user User, ctx *UseContext) bool {
	egg, ok := world.EntityByName("minecraft:egg")
	if !ok {
		return false
	}

	p, ok := egg.(interface {
		New(pos, vel mgl64.Vec3, owner world.Entity) world.Entity
	})
	if !ok {
		return false
	}

	w.PlaySound(user.Position(), sound.ItemThrow{})
	w.AddEntity(p.New(eyePosition(user), directionVector(user).Mul(1.5), user))

	ctx.SubtractFromCount(1)
	return true
}

// EncodeItem ...
func (e Egg) EncodeItem() (name string, meta int16) {
	return "minecraft:egg", 0
}
