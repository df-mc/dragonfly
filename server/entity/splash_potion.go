package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// NewSplashPotion creates a splash potion. SplashPotion is an item that grants
// effects when thrown.
func NewSplashPotion(opts world.EntitySpawnOpts, t potion.Potion, owner world.Entity) *world.EntityHandle {
	colour, _ := effect.ResultingColour(t.Effects())

	conf := splashPotionConf
	conf.Potion = t
	conf.Particle = particle.Splash{Colour: colour}
	conf.Hit = potionSplash(1, t, false)
	conf.Owner = owner.H()

	return opts.New(SplashPotionType, conf)
}

var splashPotionConf = ProjectileBehaviourConfig{
	Gravity: 0.05,
	Drag:    0.01,
	Damage:  -1,
	Sound:   sound.GlassBreak{},
}

// SplashPotionType is a world.EntityType implementation for SplashPotion.
var SplashPotionType splashPotionType

type splashPotionType struct{}

func (t splashPotionType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (splashPotionType) EncodeEntity() string { return "minecraft:splash_potion" }
func (splashPotionType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (splashPotionType) DecodeNBT(m map[string]any, data *world.EntityData) {
	conf := splashPotionConf
	conf.Potion = potion.From(nbtconv.Int32(m, "PotionId"))
	colour, _ := effect.ResultingColour(conf.Potion.Effects())
	conf.Particle = particle.Splash{Colour: colour}
	conf.Hit = potionSplash(1, conf.Potion, false)

	data.Data = conf.New()
}

func (splashPotionType) EncodeNBT(data *world.EntityData) map[string]any {
	return map[string]any{"PotionId": int32(data.Data.(*ProjectileBehaviour).conf.Potion.Uint8())}
}
