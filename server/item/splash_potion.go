package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math"
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
func (s SplashPotion) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	create := tx.World().EntityRegistry().Config().SplashPotion
	r := user.Rotation()
	r[1] = throwableOffset(r[1])
	opts := world.EntitySpawnOpts{Position: eyePosition(user), Velocity: r.Vec3().Mul(0.5)}
	tx.AddEntity(create(opts, s.Type, user))
	tx.PlaySound(user.Position(), sound.ItemThrow{})

	ctx.SubtractFromCount(1)
	return true
}

// throwableOffset adds an upwards offset pitch to a throwable entity.
// In vanilla, items such as Splash Potions, Lingering Potions, and
// Bottle o' Enchanting are thrown at a higher angle than where the
// player is looking at.
// The added offset is an ellipse-like shape based on what the input pitch is.
func throwableOffset(pitch float64) float64 {
	pitch = max(min(pitch, 89.9), -89.9)
	pitch -= math.Sqrt(math.Pow(89.9, 2)-math.Pow(pitch, 2)) * (26.5 / 89.9)
	pitch = max(min(pitch, 89.9), -89.9)

	return pitch
}

// EncodeItem ...
func (s SplashPotion) EncodeItem() (name string, meta int16) {
	return "minecraft:splash_potion", int16(s.Type.Uint8())
}
