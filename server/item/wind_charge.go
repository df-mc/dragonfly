package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"time"
)

// WindCharge is a throwable item that creates a burst of wind on impact, knocking back nearby entities and
// toggling certain blocks such as doors, trapdoors and fence gates.
type WindCharge struct{}

// Use ...
func (WindCharge) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().WindCharge
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: user.Rotation().Vec3().Mul(3.0)}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// Cooldown ...
func (WindCharge) Cooldown() time.Duration {
	return time.Millisecond * 500
}

// MaxCount ...
func (WindCharge) MaxCount() int {
	return 64
}

// EncodeItem ...
func (WindCharge) EncodeItem() (name string, meta int16) {
	return "minecraft:wind_charge", 0
}
