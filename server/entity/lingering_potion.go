package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
)

// NewLingeringPotion creates a new lingering potion. LingeringPotion is a
// variant of a splash potion that can be thrown to leave clouds with status
// effects that linger on the ground in an area.
func NewLingeringPotion(opts world.EntitySpawnOpts, t potion.Potion, owner world.Entity) *world.EntityHandle {
	colour, _ := effect.ResultingColour(t.Effects())

	conf := splashPotionConf
	conf.Potion = t
	conf.Particle = particle.Splash{Colour: colour}
	conf.Hit = potionSplash(0.25, t, true)
	conf.Owner = owner.H()
	return opts.New(LingeringPotionType, conf)
}

// LingeringPotionType is a world.EntityType implementation for LingeringPotion.
var LingeringPotionType lingeringPotionType

type lingeringPotionType struct{}

func (t lingeringPotionType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (lingeringPotionType) EncodeEntity() string {
	return "minecraft:lingering_potion"
}
func (lingeringPotionType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (lingeringPotionType) DecodeNBT(m map[string]any, data *world.EntityData) {
	conf := splashPotionConf
	conf.Potion = potion.From(nbtconv.Int32(m, "PotionId"))
	colour, _ := effect.ResultingColour(conf.Potion.Effects())
	conf.Particle = particle.Splash{Colour: colour}
	conf.Hit = potionSplash(0.25, conf.Potion, true)

	data.Data = conf.New()
}

func (lingeringPotionType) EncodeNBT(data *world.EntityData) map[string]any {
	return map[string]any{"PotionId": int32(data.Data.(*ProjectileBehaviour).conf.Potion.Uint8())}
}
