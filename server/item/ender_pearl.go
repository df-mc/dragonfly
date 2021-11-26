package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"time"
)

// EnderPearl is a smooth, greenish-blue item used to teleport and to make an eye of ender.
type EnderPearl struct{}

// CooledDown ...
func (EnderPearl) CoolDown() time.Duration {
	return time.Second
}

// Use ...
func (e EnderPearl) Use(w *world.World, user User, ctx *UseContext) bool {
	pearl, ok := world.EntityByName("minecraft:ender_pearl")
	if !ok {
		return false
	}

	p, ok := pearl.(projectile)
	if !ok {
		return false
	}

	yaw, pitch := user.Rotation()
	entity := p.New(eyePosition(user), directionVector(user).Mul(1.5), yaw, pitch)
	if o, ok := entity.(owned); ok {
		o.Own(user)
	}

	ctx.SubtractFromCount(1)

	w.PlaySound(user.Position(), sound.ItemThrow{})

	w.AddEntity(entity)

	return true
}

// MaxCount ...
func (EnderPearl) MaxCount() int {
	return 16
}

// EncodeItem ...
func (EnderPearl) EncodeItem() (name string, meta int16) {
	return "minecraft:ender_pearl", 0
}
