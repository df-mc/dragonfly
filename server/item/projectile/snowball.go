package projectile

import (
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

type Snowball struct{}

func (s Snowball) MaxCount() int {
	return 16
}

func (s Snowball) Use(w *world.World, user item.User, ctx *item.UseContext) bool {
	yaw, pitch := user.Rotation()

	var owner entity.Living
	if living, ok := user.(entity.Living); ok {
		owner = living
	}

	snow := entity.NewSnowball(entity.EyePosition(user), yaw, pitch, owner)
	snow.SetVelocity(entity.DirectionVector(user).Mul(1.5))

	ctx.SubtractFromCount(1)

	w.AddEntity(snow)

	return true
}

// EncodeItem ...
func (s Snowball) EncodeItem() (name string, meta int16) {
	return "minecraft:snowball", 0
}
