package item

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	windChargeLaunchPower     = 1.5
	windChargeLaunchDeviation = 0.0172275
)

// WindCharge is a throwable item that creates a burst of wind on impact, knocking back nearby entities and
// toggling certain blocks such as doors, trapdoors and fence gates.
type WindCharge struct{}

// Use ...
func (WindCharge) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().WindCharge
	deviation := func() float64 {
		return (rand.Float64() - rand.Float64()) * windChargeLaunchDeviation
	}
	velocity := user.Rotation().Vec3().Normalize().Add(mgl64.Vec3{
		deviation(), deviation(), deviation(),
	}).Mul(windChargeLaunchPower)

	var shooterVelocity mgl64.Vec3
	onGround := true
	if moving, ok := user.(interface {
		Velocity() mgl64.Vec3
		OnGround() bool
	}); ok {
		shooterVelocity, onGround = moving.Velocity(), moving.OnGround()
	}
	if onGround {
		shooterVelocity[1] = 0
	}
	opts := world.EntitySpawnOpts{
		Position: eyePosition(user),
		Velocity: velocity.Add(shooterVelocity),
	}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// Cooldown ...
func (WindCharge) Cooldown() time.Duration {
	return time.Millisecond * 500
}

// EncodeItem ...
func (WindCharge) EncodeItem() (name string, meta int16) {
	return "minecraft:wind_charge", 0
}
