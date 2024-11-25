package entity

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"iter"
	"time"
)

// AreaEffectCloudBehaviourConfig contains optional parameters for an area
// effect cloud entity.
type AreaEffectCloudBehaviourConfig struct {
	Potion potion.Potion
	// Radius specifies the initial radius of the cloud. Defaults to 3.0.
	Radius float64
	// RadiusUseGrowth is the value that is added to the radius every time the
	// effect cloud is used/consumed. This is typically a negative value. (-0.5)
	RadiusUseGrowth float64
	// RadiusTickGrowth is the value added to the radius every tick. This is
	// typically a negative value. (-0.005)
	RadiusTickGrowth float64
	// Duration specifies the initial duration of the cloud. Defaults to 30s.
	Duration time.Duration
	// DurationUseGrowth is the duration that is added to the effect cloud every
	// time it is used/consumed. This is 0 in normal situations.
	DurationUseGrowth time.Duration
	// ReapplicationDelay specifies the delay with which the effects from the
	// cloud can be re-applied to users.
	ReapplicationDelay time.Duration
}

func (conf AreaEffectCloudBehaviourConfig) Apply(data *world.EntityData) {
	data.Data = conf.New()
}

// New creates an AreaEffectCloudBehaviour using the parameter in conf and t.
func (conf AreaEffectCloudBehaviourConfig) New() *AreaEffectCloudBehaviour {
	if conf.Radius == 0 {
		conf.Radius = 3.0
	}
	if conf.Duration == 0 {
		conf.Duration = time.Second * 30
	}
	stationary := StationaryBehaviourConfig{ExistenceDuration: conf.Duration}
	return &AreaEffectCloudBehaviour{
		conf:       conf,
		stationary: stationary.New(),
		duration:   conf.Duration,
		radius:     conf.Radius,
		targets:    make(map[*world.EntityHandle]time.Duration),
	}
}

// AreaEffectCloudBehaviour is the cloud that is created when: lingering
// potions are thrown; creepers with potion effects explode; dragon fireballs
// hit the ground.
type AreaEffectCloudBehaviour struct {
	conf AreaEffectCloudBehaviourConfig

	stationary *StationaryBehaviour

	duration time.Duration
	radius   float64
	targets  map[*world.EntityHandle]time.Duration
}

// Radius returns the current radius of the area effect cloud.
func (a *AreaEffectCloudBehaviour) Radius() float64 {
	return a.radius
}

// Effects returns the effects the area effect cloud provides.
func (a *AreaEffectCloudBehaviour) Effects() []effect.Effect {
	return a.conf.Potion.Effects()
}

// Tick ...
func (a *AreaEffectCloudBehaviour) Tick(e *Ent, tx *world.Tx) *Movement {
	a.stationary.Tick(e, tx)
	if a.stationary.close || e.Age() < time.Second/2 {
		// The cloud lives for at least half a second before it may begin
		// spreading effects and growing/shrinking.
		return nil
	}

	pos := e.Position()
	if a.subtractTickRadius() {
		for _, v := range tx.Viewers(pos) {
			v.ViewEntityState(e)
		}
	}

	if int16(e.Age()/(time.Second*20))%10 != 0 {
		// Area effect clouds only trigger updates every ten ticks.
		return nil
	}

	for target, expiration := range a.targets {
		if e.Age() >= expiration {
			delete(a.targets, target)
		}
	}
	if a.applyEffects(pos, e, a.filter(tx.EntitiesWithin(e.H().Type().BBox(e).Translate(pos)))) {
		for _, v := range tx.Viewers(pos) {
			v.ViewEntityState(e)
		}
	}
	return nil
}

func (a *AreaEffectCloudBehaviour) filter(seq iter.Seq[world.Entity]) iter.Seq[world.Entity] {
	return func(yield func(world.Entity) bool) {
		for e := range seq {
			_, target := a.targets[e.H()]
			_, living := e.(Living)
			if !living || target {
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
}

// applyEffects applies the effects of an area effect cloud at pos to all
// entities passed if they were within the radius and don't have an active
// cooldown period.
func (a *AreaEffectCloudBehaviour) applyEffects(pos mgl64.Vec3, ent *Ent, entities iter.Seq[world.Entity]) bool {
	var update bool
	for e := range entities {
		delta := e.Position().Sub(pos)
		delta[1] = 0
		if delta.Len() <= a.radius {
			l := e.(Living)
			for _, eff := range a.Effects() {
				if lasting, ok := eff.Type().(effect.LastingType); ok {
					l.AddEffect(effect.New(lasting, eff.Level(), eff.Duration()/4))
					continue
				}
				l.AddEffect(eff)
			}

			a.targets[e.H()] = ent.Age() + a.conf.ReapplicationDelay
			a.subtractUseDuration()
			a.subtractUseRadius()

			update = true
		}
	}
	return update
}

// subtractTickRadius grows the cloud's radius by the radiusTickGrowth value. If the
// radius goes under 1/2, it will close the entity.
func (a *AreaEffectCloudBehaviour) subtractTickRadius() bool {
	a.radius += a.conf.RadiusTickGrowth
	if a.radius < 0.5 {
		a.stationary.close = true
	}
	return a.conf.RadiusTickGrowth != 0
}

// subtractUseDuration grows duration by the durationUseGrowth factor. If duration
// goes under zero, it will close the entity.
func (a *AreaEffectCloudBehaviour) subtractUseDuration() {
	a.duration += a.conf.DurationUseGrowth
	if a.duration <= 0 {
		a.stationary.close = true
	}
}

// subtractUseRadius grows radius by the radiusUseGrowth factor. If radius goes
// under 1/2, it will close the entity.
func (a *AreaEffectCloudBehaviour) subtractUseRadius() {
	a.radius += a.conf.RadiusUseGrowth
	if a.radius <= 0.5 {
		a.stationary.close = true
	}
}
