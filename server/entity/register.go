package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/tools/go/cfg"
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
	Item: func(opts world.EntitySpawnOpts, it any) *world.EntityHandle {
		i := NewItem(it.(item.Stack), pos)
		i.vel = vel
		return i
	},
	FallingBlock: func(bl world.Block, pos mgl64.Vec3) *world.EntityHandle {
		return NewFallingBlock(bl, pos)
	},
	TNT: func(pos mgl64.Vec3, fuse time.Duration, igniter world.Entity) *world.EntityHandle {
		return NewTNT(pos, fuse, igniter)
	},
	BottleOfEnchanting: func(pos, vel mgl64.Vec3, owner world.Entity) *world.EntityHandle {
		b := NewBottleOfEnchanting(pos, owner)
		b.vel = vel
		return b
	},
	Arrow: func(opts world.EntitySpawnOpts, damage float64, owner world.Entity, critical, disallowPickup, obtainArrowOnPickup bool, punchLevel int, tip any) *world.EntityHandle {
		conf := arrowConf
		conf.Damage = damage
		conf.Potion = tip.(potion.Potion)
		conf.Owner = owner
		conf.KnockBackForceAddend = float64(punchLevel) * (enchantment.Punch{}).KnockBackMultiplier()
		conf.DisablePickup = disallowPickup
		if obtainArrowOnPickup {
			conf.PickupItem = item.NewStack(item.Arrow{Tip: tip.(potion.Potion)}, 1)
		}
		conf.Critical = critical
		return cfg.New(ArrowType{}, conf)
	},
	Egg: func(pos, vel mgl64.Vec3, owner world.Entity) *world.EntityHandle {
		e := NewEgg(pos, owner)
		e.vel = vel
		return e
	},
	EnderPearl: func(pos, vel mgl64.Vec3, owner world.Entity) *world.EntityHandle {
		e := NewEnderPearl(pos, owner)
		e.vel = vel
		return e
	},
	Firework: func(pos mgl64.Vec3, rot cube.Rotation, attached bool, firework world.Item, owner world.Entity) *world.EntityHandle {
		return NewFireworkAttached(pos, rot, firework.(item.Firework), owner, attached)
	},
	LingeringPotion: func(opts world.EntitySpawnOpts, t any, owner world.Entity) *world.EntityHandle {
		return NewLingeringPotion(opts, t.(potion.Potion), owner)
	},
	Snowball: func(pos, vel mgl64.Vec3, owner world.Entity) *world.EntityHandle {
		s := NewSnowball(pos, owner)
		s.vel = vel
		return s
	},
	SplashPotion: func(pos, vel mgl64.Vec3, t any, owner world.Entity) *world.EntityHandle {
		p := NewSplashPotion(pos, owner, t.(potion.Potion))
		p.vel = vel
		return p
	},
	Lightning: NewLightning,
}
