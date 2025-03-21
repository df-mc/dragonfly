package playerdb

import "github.com/df-mc/dragonfly/server/entity/effect"

func effectsToData(effects []effect.Effect) []jsonEffect {
	data := make([]jsonEffect, len(effects))
	for key, eff := range effects {
		id, ok := effect.ID(eff.Type())
		if !ok {
			continue
		}
		data[key] = jsonEffect{
			ID:              id,
			Duration:        eff.Duration(),
			Level:           eff.Level(),
			Ambient:         eff.Ambient(),
			ParticlesHidden: eff.ParticlesHidden(),
			Infinite:        eff.Infinite(),
		}
	}
	return data
}

func dataToEffects(data []jsonEffect) []effect.Effect {
	effects := make([]effect.Effect, len(data))
	for i, d := range data {
		e, ok := effect.ByID(d.ID)
		if !ok {
			continue
		}
		switch eff := e.(type) {
		case effect.LastingType:
			if d.Ambient {
				effects[i] = effect.NewAmbient(eff, d.Level, d.Duration)
			} else if d.Infinite {
				effects[i] = effect.NewInfinite(eff, d.Level)
			} else {
				effects[i] = effect.New(eff, d.Level, d.Duration)
			}

			if d.ParticlesHidden {
				effects[i] = effects[i].WithoutParticles()
			}
		default:
			effects[i] = effect.NewInstant(eff, d.Level)
		}
	}
	return effects
}
