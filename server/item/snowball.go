package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Snowball is a throwable combat item obtained through shovelling snow.
type Snowball struct{}

// MaxCount ...
func (s Snowball) MaxCount() int {
	return 16
}

// Use ...
func (s Snowball) Use(w *world.World, user User, ctx *UseContext) bool {
	snow, ok := world.EntityByName("minecraft:snowball")
	if !ok {
		return false
	}

	p, ok := snow.(projectile)
	if !ok {
		return false
	}

	yaw, pitch := user.Rotation()
	e := p.New(eyePosition(user), directionVector(user).Mul(1.5), yaw, pitch)
	if o, ok := e.(owned); ok {
		o.Own(user)
	}

	ctx.SubtractFromCount(1)

	w.PlaySound(user.Position(), sound.ItemThrow{})

	w.AddEntity(e)

	return true
}

// EncodeItem ...
func (s Snowball) EncodeItem() (name string, meta int16) {
	return "minecraft:snowball", 0
}
