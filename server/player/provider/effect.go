package provider

import "github.com/df-mc/dragonfly/server/entity/effect"

func effectsToData(effects []effect.Effect) []jsonEffect {
	data := make([]jsonEffect, len(effects))
	for key, eff := range effects {
		id, ok := effect.ID(eff)
		if !ok {
			continue
		}
		data[key] = jsonEffect{
			ID:       id,
			Duration: eff.Duration(),
			Level:    eff.Level(),
			Ambient:  eff.AmbientSource(),
		}
	}
	return data
}

func dataToEffects(data []jsonEffect) []effect.Effect {
	effects := make([]effect.Effect, len(data))
	for key, d := range data {
		e, ok := effect.ByID(d.ID)
		if !ok {
			continue
		}
		effects[key] = e.WithSettings(d.Duration, d.Level, d.Ambient)
	}
	return effects
}
