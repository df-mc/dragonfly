package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// NewAreaEffectCloud creates a new area effect cloud entity and returns it.
func NewAreaEffectCloud(opts world.EntitySpawnOpts, p potion.Potion) *world.EntityHandle {
	config := areaEffectCloudConf
	config.Potion = p
	for _, e := range p.Effects() {
		if _, ok := e.Type().(effect.LastingType); !ok {
			config.ReapplicationDelay = 0
			break
		}
	}
	return opts.New(AreaEffectCloudType, config)
}

var areaEffectCloudConf = AreaEffectCloudBehaviourConfig{
	RadiusUseGrowth:    -0.5,
	RadiusTickGrowth:   -0.005,
	ReapplicationDelay: time.Second * 2,
}

// NewAreaEffectCloudWith ...
func NewAreaEffectCloudWith(opts world.EntitySpawnOpts, t potion.Potion, duration, reapplicationDelay, durationOnUse time.Duration, radius, radiusOnUse, radiusGrowth float64) *world.EntityHandle {
	config := AreaEffectCloudBehaviourConfig{
		Potion:             t,
		Radius:             radius,
		RadiusUseGrowth:    radiusOnUse,
		RadiusTickGrowth:   radiusGrowth,
		Duration:           duration,
		DurationUseGrowth:  durationOnUse,
		ReapplicationDelay: reapplicationDelay,
	}
	return opts.New(AreaEffectCloudType, config)
}

// AreaEffectCloudType is a world.EntityType implementation for AreaEffectCloud.
var AreaEffectCloudType areaEffectCloudType

type areaEffectCloudType struct{}

func (t areaEffectCloudType) Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) world.Entity {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (areaEffectCloudType) EncodeEntity() string { return "minecraft:area_effect_cloud" }
func (areaEffectCloudType) BBox(e world.Entity) cube.BBox {
	r := e.(*Ent).Behaviour().(*AreaEffectCloudBehaviour).Radius()
	return cube.Box(-r, 0, -r, r, 0.5, r)
}

func (areaEffectCloudType) DecodeNBT(m map[string]any, data *world.EntityData) {
	data.Data = AreaEffectCloudBehaviourConfig{
		Potion:             potion.From(nbtconv.Int32(m, "PotionId")),
		Radius:             float64(nbtconv.Float32(m, "Radius")),
		RadiusUseGrowth:    float64(nbtconv.Float32(m, "RadiusOnUse")),
		RadiusTickGrowth:   float64(nbtconv.Float32(m, "RadiusPerTick")),
		Duration:           nbtconv.TickDuration[int32](m, "Duration"),
		DurationUseGrowth:  nbtconv.TickDuration[int32](m, "ReapplicationDelay"),
		ReapplicationDelay: nbtconv.TickDuration[int32](m, "DurationOnUse"),
	}.New()
}

func (areaEffectCloudType) EncodeNBT(data *world.EntityData) map[string]any {
	a := data.Data.(*AreaEffectCloudBehaviour)
	return map[string]any{
		"PotionId":           int32(a.conf.Potion.Uint8()),
		"ReapplicationDelay": int32(a.conf.ReapplicationDelay),
		"RadiusPerTick":      float32(a.conf.RadiusTickGrowth),
		"RadiusOnUse":        float32(a.conf.RadiusUseGrowth),
		"DurationOnUse":      int32(a.conf.DurationUseGrowth),
		"Radius":             float32(a.radius),
		"Duration":           int32(a.duration),
	}
}
