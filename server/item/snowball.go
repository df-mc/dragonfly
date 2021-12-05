package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
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

	p, ok := snow.(interface {
		New(pos, vel mgl64.Vec3, yaw, pitch float64) world.Entity
	})
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
