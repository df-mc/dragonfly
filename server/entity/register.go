package entity

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// DefaultRegistry is a world.EntityRegistry that registers all default entities
// implemented by Dragonfly.
var DefaultRegistry = conf.New([]world.EntityType{
	AreaEffectCloudType{},
	ArrowType{},
	BottleOfEnchantingType{},
	EggType{},
	EnderPearlType{},
	ExperienceOrbType{},
	FallingBlockType{},
	FireworkType{},
	ItemType{},
	LightningType{},
	LingeringPotionType{},
	SnowballType{},
	SplashPotionType{},
	TNTType{},
	TextType{},
})

var conf = world.EntityRegistryConfig{
	Item: func(it any, pos, vel mgl64.Vec3) world.Entity {
		i := NewItem(it.(item.Stack), pos)
		i.vel = vel
		return i
	},
	FallingBlock: func(bl world.Block, pos mgl64.Vec3) world.Entity {
		return NewFallingBlock(bl, pos)
	},
	TNT: func(pos mgl64.Vec3, fuse time.Duration) world.Entity {
		return NewTNT(pos, fuse)
	},
	BottleOfEnchanting: func(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
		b := NewBottleOfEnchanting(pos, owner)
		b.vel = vel
		return b
	},
	Arrow: func(pos, vel mgl64.Vec3, yaw, pitch, damage float64, owner world.Entity, critical, disallowPickup, obtainArrowOnPickup bool, punchLevel int, tip any) world.Entity {
		a := NewTippedArrowWithDamage(pos, yaw, pitch, damage, owner, tip.(potion.Potion))
		b := a.conf.Behaviour.(*ProjectileBehaviour)
		b.conf.KnockBackAddend = float64(punchLevel) * (enchantment.Punch{}).KnockBackMultiplier()
		b.conf.DisablePickup = disallowPickup
		if obtainArrowOnPickup {
			b.conf.PickupItem = item.NewStack(item.Arrow{Tip: tip.(potion.Potion)}, 1)
		}
		b.conf.Critical = critical
		a.vel = vel
		return a
	},
	Egg: func(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
		e := NewEgg(pos, owner)
		e.vel = vel
		return e
	},
	EnderPearl: func(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
		e := NewEnderPearl(pos, owner)
		e.vel = vel
		return e
	},
	Firework: func(pos mgl64.Vec3, yaw, pitch float64, attached bool, firework world.Item, owner world.Entity) world.Entity {
		f := NewFirework(pos, yaw, pitch, firework.(item.Firework))
		f.owner = owner
		f.attached = attached
		return f
	},
	LingeringPotion: func(pos, vel mgl64.Vec3, t any, owner world.Entity) world.Entity {
		p := NewLingeringPotion(pos, owner, t.(potion.Potion))
		p.vel = vel
		return p
	},
	Snowball: func(pos, vel mgl64.Vec3, owner world.Entity) world.Entity {
		s := NewSnowball(pos, owner)
		s.vel = vel
		return s
	},
	SplashPotion: func(pos, vel mgl64.Vec3, t any, owner world.Entity) world.Entity {
		p := NewSplashPotion(pos, owner, t.(potion.Potion))
		p.vel = vel
		return p
	},
	Lightning: func(pos mgl64.Vec3) world.Entity {
		return NewLightning(pos)
	},
}
