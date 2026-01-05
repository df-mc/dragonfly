package entity

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
)

// DefaultRegistry is a world.EntityRegistry that registers all default entities
// implemented by Dragonfly.
var DefaultRegistry = conf.New([]world.EntityType{
	AreaEffectCloudType,
	ArrowType,
	BottleOfEnchantingType,
	EggType,
	EnderPearlType,
	ExperienceOrbType,
	FallingBlockType,
	FireworkType,
	ItemType,
	LightningType,
	LingeringPotionType,
	SnowballType,
	SplashPotionType,
	TNTType,
	TextType,
	WindChargeType,
})

var conf = world.EntityRegistryConfig{
	TNT:                NewTNT,
	Egg:                NewEgg,
	Snowball:           NewSnowball,
	BottleOfEnchanting: NewBottleOfEnchanting,
	EnderPearl:         NewEnderPearl,
	FallingBlock:       NewFallingBlock,
	Lightning:          NewLightning,
	WindCharge:         NewWindCharge,
	Firework: func(opts world.EntitySpawnOpts, firework world.Item, owner world.Entity, sidewaysVelocityMultiplier, upwardsAcceleration float64, attached bool) *world.EntityHandle {
		return newFirework(opts, firework.(item.Firework), owner, sidewaysVelocityMultiplier, upwardsAcceleration, attached)
	},
	Item: func(opts world.EntitySpawnOpts, it any) *world.EntityHandle {
		return NewItem(opts, it.(item.Stack))
	},
	LingeringPotion: func(opts world.EntitySpawnOpts, t any, owner world.Entity) *world.EntityHandle {
		return NewLingeringPotion(opts, t.(potion.Potion), owner)
	},
	SplashPotion: func(opts world.EntitySpawnOpts, t any, owner world.Entity) *world.EntityHandle {
		return NewSplashPotion(opts, t.(potion.Potion), owner)
	},
	Arrow: func(opts world.EntitySpawnOpts, damage float64, owner world.Entity, critical, disallowPickup, obtainArrowOnPickup bool, punchLevel int, tip any) *world.EntityHandle {
		conf := arrowConf
		conf.Damage, conf.Potion, conf.Owner = damage, tip.(potion.Potion), owner.H()
		conf.KnockBackForceAddend = float64(punchLevel) * enchantment.Punch.KnockBackMultiplier()
		conf.DisablePickup = disallowPickup
		if obtainArrowOnPickup {
			conf.PickupItem = item.NewStack(item.Arrow{Tip: tip.(potion.Potion)}, 1)
		}
		conf.Critical = critical
		return opts.New(ArrowType, conf)
	},
}
