package item

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// FishingRod is a tool used to catch fish and other items from bodies of water. It may also be used to pull
// entities towards the user.
type FishingRod struct{}

// Angler represents a User that is able to cast a fishing rod. It keeps track of the fishing hook it has
// cast, so that the hook may be reeled in when the fishing rod is used again.
type Angler interface {
	User
	// FishingHook returns the handle of the fishing hook currently cast by the Angler, or nil if the
	// Angler does not have a fishing hook cast.
	FishingHook() *world.EntityHandle
	// SetFishingHook changes the fishing hook cast by the Angler. Nil may be passed to clear it.
	SetFishingHook(hook *world.EntityHandle)
	// ReelFishingHook reels in the fishing hook currently cast by the Angler. It returns the durability
	// damage that the fishing rod should take and true, or false if the Angler had no fishing hook cast.
	ReelFishingHook() (damage int, ok bool)
}

// MaxCount always returns 1.
func (FishingRod) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (FishingRod) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 384,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// FuelInfo ...
func (FishingRod) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// EnchantmentValue ...
func (FishingRod) EnchantmentValue() int {
	return 1
}

// Use casts a fishing hook if the user does not yet have one cast, or reels in the fishing hook that was
// previously cast.
func (FishingRod) Use(tx *world.Tx, user User, ctx *UseContext) bool {
	angler, ok := user.(Angler)
	if !ok {
		return false
	}
	if damage, ok := angler.ReelFishingHook(); ok {
		ctx.DamageItem(damage)
		return true
	}

	held, _ := user.HeldItems()
	lure, luck := 0, 0
	for _, enchant := range held.Enchantments() {
		if l, ok := enchant.Type().(interface{ LureLevel(int) int }); ok {
			lure = l.LureLevel(enchant.Level())
		}
		if l, ok := enchant.Type().(interface{ FishingLuck(int) int }); ok {
			luck = l.FishingLuck(enchant.Level())
		}
	}

	pos, vel := fishingHookLaunch(user)
	create := tx.World().EntityRegistry().Config().FishingHook
	hook := tx.AddEntity(create(world.EntitySpawnOpts{Position: pos, Velocity: vel}, user, lure, luck))
	angler.SetFishingHook(hook.H())
	tx.PlaySound(user.Position(), sound.ItemThrow{})
	return true
}

// fishingHookLaunch returns the position and velocity with which a fishing hook is launched by the user
// passed, matching the launch of a fishing hook in vanilla.
func fishingHookLaunch(user User) (pos, vel mgl64.Vec3) {
	yaw, pitch := mgl64.DegToRad(user.Rotation().Yaw()), mgl64.DegToRad(user.Rotation().Pitch())
	horizontal := mgl64.Vec3{-math.Sin(yaw), 0, math.Cos(yaw)}

	pos = eyePosition(user).Add(horizontal.Mul(0.3))
	vel = mgl64.Vec3{horizontal[0], math.Max(math.Min(-math.Tan(pitch), 5), -5), horizontal[2]}
	l := vel.Len()
	for i := range vel {
		vel[i] *= 0.6/l + 0.5 + 0.0103365*(rand.Float64()-rand.Float64())
	}
	return pos, vel
}

// EncodeItem ...
func (FishingRod) EncodeItem() (name string, meta int16) {
	return "minecraft:fishing_rod", 0
}
