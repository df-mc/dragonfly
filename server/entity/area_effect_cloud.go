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

// AreaEffectCloud is the cloud that is created when: lingering potions are thrown; creepers with potion effects explode;
// dragon fireballs hit the ground.
type AreaEffectCloud struct {
	duration           int64
	reapplicationDelay int64
	durationOnUse      int64

	radius       float64
	radiusOnUse  float64
	radiusGrowth float64

	close bool

	targets map[world.Entity]int64

	t   potion.Potion
	age int64

	transform
}

// NewDefaultAreaEffectCloud ...
func NewDefaultAreaEffectCloud(pos mgl64.Vec3, t potion.Potion) *AreaEffectCloud {
	return NewAreaEffectCloud(pos, t, time.Second*30, time.Second*2, 0, 3.0, -0.5, -0.005)
}

// NewAreaEffectCloud ...
func NewAreaEffectCloud(pos mgl64.Vec3, t potion.Potion, duration, reapplicationDelay, durationOnUse time.Duration, radius, radiusOnUse, radiusGrowth float64) *AreaEffectCloud {
	a := &AreaEffectCloud{
		duration:           duration.Milliseconds() / 50,
		reapplicationDelay: reapplicationDelay.Milliseconds() / 50,
		durationOnUse:      durationOnUse.Milliseconds() / 50,

		radius:       radius,
		radiusOnUse:  radiusOnUse,
		radiusGrowth: radiusGrowth,

		targets: make(map[world.Entity]int64),
		t:       t,
	}
	a.transform = newTransform(a, pos)
	return a
}

// Name ...
func (a *AreaEffectCloud) Name() string {
	return "Area Effect Cloud"
}

// EncodeEntity ...
func (a *AreaEffectCloud) EncodeEntity() string {
	return "minecraft:area_effect_cloud"
}

// Duration returns the duration of the cloud.
func (a *AreaEffectCloud) Duration() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	return time.Duration(a.duration) * time.Millisecond * 50
}

// Radius returns information about the cloud's, radius, change rate, and growth rate.
func (a *AreaEffectCloud) Radius() (radius, radiusOnUse, radiusGrowth float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.radius, a.radiusOnUse, a.radiusGrowth
}

// Effects returns the effects the area effect cloud provides.
func (a *AreaEffectCloud) Effects() []effect.Effect {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.t.Effects()
}

// BBox ...
func (a *AreaEffectCloud) BBox() cube.BBox {
	a.mu.Lock()
	defer a.mu.Unlock()
	return cube.Box(-a.radius, 0, -a.radius, a.radius, 0.5, a.radius)
}

// Tick ...
func (a *AreaEffectCloud) Tick(w *world.World, _ int64) {
	if a.close {
		_ = a.Close()
		return
	}

	a.age++
	if a.age < 10 {
		// The cloud lives for at least half a second before it may begin spreading effects and growing/shrinking.
		return
	}

	a.mu.Lock()
	if a.age >= a.duration+10 {
		// We've outlived our duration, close the entity the next tick.
		a.close = true
		a.mu.Unlock()
		return
	}

	age, growth := a.age, a.radiusGrowth
	pos := a.pos
	a.mu.Unlock()

	if growth != 0 {
		a.mu.Lock()
		a.radius += growth
		if a.radius < 0.5 {
			a.close = true
			a.mu.Unlock()
			return
		}
		a.mu.Unlock()

		for _, v := range w.Viewers(pos) {
			v.ViewEntityState(a)
		}
	}

	if age%5 != 0 {
		// Area effect clouds only trigger updates every five ticks.
		return
	}

	for target, expiration := range a.targets {
		if a.age >= expiration {
			delete(a.targets, target)
		}
	}

	entities := w.EntitiesWithin(a.BBox().Translate(pos), func(entity world.Entity) bool {
		_, target := a.targets[entity]
		_, living := entity.(Living)
		return !living || target || entity == a
	})

	a.mu.Lock()

	var updated bool
	for _, e := range entities {
		delta := e.Position().Sub(pos)
		delta[1] = 0
		if delta.Len() <= a.radius {
			l := e.(Living)
			for _, eff := range a.t.Effects() {
				if lasting, ok := eff.Type().(effect.LastingType); ok {
					l.AddEffect(effect.New(lasting, eff.Level(), eff.Duration()/4))
					continue
				}
				l.AddEffect(eff)
			}

			a.targets[e] = a.age + a.reapplicationDelay
			if a.radiusOnUse != 0.0 {
				a.radius += a.radiusOnUse
				updated = true
				if a.radius < 0.5 {
					a.close = true
					a.mu.Unlock()
					return
				}
			}
			if a.durationOnUse != 0 {
				a.duration += a.durationOnUse
				updated = true
				if a.duration <= 0 {
					a.close = true
					a.mu.Unlock()
					return
				}
			}
		}
	}
	a.mu.Unlock()

	if updated {
		for _, v := range w.Viewers(pos) {
			v.ViewEntityState(a)
		}
	}
}

// EncodeNBT ...
func (a *AreaEffectCloud) EncodeNBT() map[string]any {
	return map[string]any{
		"Pos":                nbtconv.Vec3ToFloat32Slice(a.Position()),
		"ReapplicationDelay": int32(a.reapplicationDelay),
		"RadiusPerTick":      float32(a.radiusGrowth),
		"RadiusOnUse":        float32(a.radiusOnUse),
		"DurationOnUse":      int32(a.durationOnUse),
		"Radius":             float32(a.radius),
		"Duration":           int32(a.duration),
		"PotionId":           a.t.Uint8(),
	}
}

// DecodeNBT ...
func (a *AreaEffectCloud) DecodeNBT(data map[string]any) any {
	return NewAreaEffectCloud(
		nbtconv.MapVec3(data, "Pos"),
		potion.From(nbtconv.Map[int32](data, "PotionId")),
		time.Duration(nbtconv.Map[int32](data, "Duration"))*time.Millisecond*50,
		time.Duration(nbtconv.Map[int32](data, "ReapplicationDelay"))*time.Millisecond*50,
		time.Duration(nbtconv.Map[int32](data, "DurationOnUse"))*time.Millisecond*50,
		float64(nbtconv.Map[float32](data, "Radius")),
		float64(nbtconv.Map[float32](data, "RadiusOnUse")),
		float64(nbtconv.Map[float32](data, "RadiusPerTick")),
	)
}
