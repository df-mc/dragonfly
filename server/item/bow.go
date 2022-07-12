package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"time"
)

// Bow is a ranged weapon that fires arrows.
type Bow struct{}

// MaxCount always returns 1.
func (Bow) MaxCount() int {
	return 1
}

// DurabilityInfo ...
func (Bow) DurabilityInfo() DurabilityInfo {
	return DurabilityInfo{
		MaxDurability: 385,
		BrokenItem:    simpleItem(Stack{}),
	}
}

// Release ...
func (Bow) Release(releaser Releaser, duration time.Duration, ctx *UseContext) {
	creative := releaser.GameMode().CreativeInventory()
	ticks := duration.Milliseconds() / 50
	if ticks < 3 {
		return
	}

	d := float64(ticks) / 20
	force := math.Min((d*d+d*2)/3, 1)
	if force < 0.1 {
		return
	}

	var arrow Stack
	var ok bool
	if arrow, ok = ctx.FirstFunc(func(stack Stack) bool {
		_, ok := stack.Item().(Arrow)
		return ok
	}); !ok && !creative {
		return
	}

	rYaw, rPitch := releaser.Rotation()
	yaw, pitch := -rYaw, -rPitch
	if rYaw > 180 {
		yaw = 360 - rYaw
	}

	proj, ok := world.EntityByName("minecraft:arrow")
	if !ok {
		return
	}

	if p, ok := proj.(interface {
		New(pos, vel mgl64.Vec3, yaw, pitch, damage float64, owner world.Entity, critical, disallowPickup, obtainArrowOnPickup bool, punchLevel int, tip potion.Potion) world.Entity
	}); ok {
		tip := arrow.Item().(Arrow).Tip
		held, _ := releaser.HeldItems()

		damage, punchLevel, burnDuration, consume := 2.0, 0, time.Duration(0), true
		for _, enchant := range held.Enchantments() {
			if f, ok := enchant.Type().(interface {
				BurnDuration() time.Duration
			}); ok {
				burnDuration = f.BurnDuration()
			}
			if _, ok := enchant.Type().(interface {
				PunchMultiplier(level int, horizontalSpeed float64) float64
			}); ok {
				punchLevel = enchant.Level()
			}
			if p, ok := enchant.Type().(interface {
				PowerDamage(level int) float64
			}); ok {
				damage += p.PowerDamage(enchant.Level())
			}
			if i, ok := enchant.Type().(interface {
				ConsumesArrows() bool
			}); ok && !i.ConsumesArrows() {
				consume = false
			}
		}

		projectile := p.New(eyePosition(releaser), directionVector(releaser).Mul(force*3), yaw, pitch, damage, releaser, force >= 1, false, !creative && consume, punchLevel, tip)
		if f, ok := projectile.(interface {
			SetOnFire(duration time.Duration)
		}); ok {
			f.SetOnFire(burnDuration)
		}

		ctx.DamageItem(1)
		if consume {
			ctx.Consume(arrow.Grow(-arrow.Count() + 1))
		}

		releaser.PlaySound(sound.BowShoot{})
		releaser.World().AddEntity(projectile)
	}
}

// Requirements returns the required items to release this item.
func (Bow) Requirements() []Stack {
	return []Stack{NewStack(Arrow{}, 1)}
}

// EncodeItem ...
func (Bow) EncodeItem() (name string, meta int16) {
	return "minecraft:bow", 0
}
