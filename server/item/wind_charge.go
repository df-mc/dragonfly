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
	var shooterVelocity mgl64.Vec3
	onGround := true
	if moving, ok := user.(interface {
		Velocity() mgl64.Vec3
		OnGround() bool
	}); ok {
		shooterVelocity, onGround = moving.Velocity(), moving.OnGround()
	}
	deviation := mgl64.Vec3{
		(rand.Float64() - rand.Float64()) * windChargeLaunchDeviation,
		(rand.Float64() - rand.Float64()) * windChargeLaunchDeviation,
		(rand.Float64() - rand.Float64()) * windChargeLaunchDeviation,
	}
	opts := world.EntitySpawnOpts{
		Position: eyePosition(user),
		Velocity: windChargeLaunchVelocity(user.Rotation().Vec3(), deviation, shooterVelocity, onGround),
	}
	tx.AddEntity(create(opts, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

func windChargeLaunchVelocity(direction, deviation, shooterVelocity mgl64.Vec3, onGround bool) mgl64.Vec3 {
	velocity := direction.Normalize().Add(deviation).Mul(windChargeLaunchPower)
	if onGround {
		shooterVelocity[1] = 0
	}
	return velocity.Add(shooterVelocity)
}

// Cooldown ...
func (WindCharge) Cooldown() time.Duration {
	return time.Millisecond * 500
}

// EncodeItem ...
func (WindCharge) EncodeItem() (name string, meta int16) {
	return "minecraft:wind_charge", 0
}
