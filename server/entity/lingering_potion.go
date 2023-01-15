package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// NewLingeringPotion creates a new lingering potion. LingeringPotion is a
// variant of a splash potion that can be thrown to leave clouds with status
// effects that linger on the ground in an area.
func NewLingeringPotion(pos mgl64.Vec3, owner world.Entity, t potion.Potion) *Ent {
	colour, _ := effect.ResultingColour(t.Effects())

	conf := splashPotionConf
	conf.Potion = t
	conf.Particle = particle.Splash{Colour: colour}
	conf.Hit = potionSplash(0.25, t, true)
	return Config{Behaviour: conf.New(owner)}.New(LingeringPotionType{}, pos)
}

// LingeringPotionType is a world.EntityType implementation for LingeringPotion.
type LingeringPotionType struct{}

func (LingeringPotionType) EncodeEntity() string {
	return "minecraft:lingering_potion"
}
func (LingeringPotionType) Glint() bool { return true }
func (LingeringPotionType) BBox(world.Entity) cube.BBox {
	return cube.Box(-0.125, 0, -0.125, 0.125, 0.25, 0.125)
}

func (LingeringPotionType) DecodeNBT(m map[string]any) world.Entity {
	pot := NewLingeringPotion(nbtconv.Vec3(m, "Pos"), nil, potion.From(nbtconv.Int32(m, "PotionId")))
	pot.vel = nbtconv.Vec3(m, "Motion")
	return pot
}

func (LingeringPotionType) EncodeNBT(e world.Entity) map[string]any {
	pot := e.(*Ent)
	return map[string]any{
		"Pos":      nbtconv.Vec3ToFloat32Slice(pot.Position()),
		"Motion":   nbtconv.Vec3ToFloat32Slice(pot.Velocity()),
		"PotionId": int32(pot.conf.Behaviour.(*ProjectileBehaviour).conf.Potion.Uint8()),
	}
}
