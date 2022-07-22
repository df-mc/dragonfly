package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// LingeringPotion is a variant of a splash potion that can be thrown to leave clouds with status effects that linger on
// the ground in an area.
type LingeringPotion struct {
	// Type is the type of lingering potion.
	Type potion.Potion
}

// MaxCount ...
func (l LingeringPotion) MaxCount() int {
	return 1
}

// Use ...
func (l LingeringPotion) Use(w *world.World, user User, ctx *UseContext) bool {
	lingering, ok := world.EntityByName("minecraft:lingering_potion")
	if !ok {
		return false
	}

	p, ok := lingering.(interface {
		New(pos, vel mgl64.Vec3, yaw, pitch float64, t potion.Potion) world.Entity
	})
	if !ok {
		return false
	}

	yaw, pitch := user.Rotation()
	e := p.New(eyePosition(user), directionVector(user).Mul(0.5), yaw, pitch, l.Type)
	if o, ok := e.(owned); ok {
		o.Own(user)
	}

	ctx.SubtractFromCount(1)

	w.PlaySound(user.Position(), sound.ItemThrow{})
	w.AddEntity(e)
	return true
}

// EncodeItem ...
func (l LingeringPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:lingering_potion", int16(l.Type.Uint8())
}
