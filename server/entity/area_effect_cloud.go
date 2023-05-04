package entity

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// NewAreaEffectCloud creates a new area effect cloud entity and returns it.
func NewAreaEffectCloud(pos mgl64.Vec3, p potion.Potion) *Ent {
	config := areaEffectCloudConf
	for _, e := range p.Effects() {
		if _, ok := e.Type().(effect.LastingType); !ok {
			config.ReapplicationDelay = 0
			break
		}
	}
	return Config{Behaviour: config.New(p)}.New(AreaEffectCloudType{}, pos)
}

var areaEffectCloudConf = AreaEffectCloudBehaviourConfig{
	RadiusUseGrowth:    -0.5,
	RadiusTickGrowth:   -0.005,
	ReapplicationDelay: time.Second * 2,
}

// NewAreaEffectCloudWith ...
func NewAreaEffectCloudWith(pos mgl64.Vec3, t potion.Potion, duration, reapplicationDelay, durationOnUse time.Duration, radius, radiusOnUse, radiusGrowth float64) *Ent {
	config := AreaEffectCloudBehaviourConfig{
		Radius:             radius,
		RadiusUseGrowth:    radiusOnUse,
		RadiusTickGrowth:   radiusGrowth,
		Duration:           duration,
		DurationUseGrowth:  durationOnUse,
		ReapplicationDelay: reapplicationDelay,
	}
	return Config{Behaviour: config.New(t)}.New(AreaEffectCloudType{}, pos)
}

// AreaEffectCloudType is a world.EntityType implementation for AreaEffectCloud.
type AreaEffectCloudType struct{}

func (AreaEffectCloudType) EncodeEntity() string { return "minecraft:area_effect_cloud" }
func (AreaEffectCloudType) BBox(e world.Entity) cube.BBox {
	r := e.(*Ent).Behaviour().(*AreaEffectCloudBehaviour).Radius()
	return cube.Box(-r, 0, -r, r, 0.5, r)
}

func (AreaEffectCloudType) DecodeNBT(m map[string]any) world.Entity {
	return NewAreaEffectCloudWith(
		nbtconv.Vec3(m, "Pos"),
		potion.From(nbtconv.Int32(m, "PotionId")),
		nbtconv.TickDuration[int32](m, "Duration"),
		nbtconv.TickDuration[int32](m, "ReapplicationDelay"),
		nbtconv.TickDuration[int32](m, "DurationOnUse"),
		float64(nbtconv.Float32(m, "Radius")),
		float64(nbtconv.Float32(m, "RadiusOnUse")),
		float64(nbtconv.Float32(m, "RadiusPerTick")),
	)
}

func (AreaEffectCloudType) EncodeNBT(e world.Entity) map[string]any {
	ent := e.(*Ent)
	a := ent.Behaviour().(*AreaEffectCloudBehaviour)
	return map[string]any{
		"Pos":                nbtconv.Vec3ToFloat32Slice(ent.Position()),
		"ReapplicationDelay": int32(a.conf.ReapplicationDelay),
		"RadiusPerTick":      float32(a.conf.RadiusTickGrowth),
		"RadiusOnUse":        float32(a.conf.RadiusUseGrowth),
		"DurationOnUse":      int32(a.conf.DurationUseGrowth),
		"Radius":             float32(a.radius),
		"Duration":           int32(a.duration),
		"PotionId":           int32(a.t.Uint8()),
	}
}
