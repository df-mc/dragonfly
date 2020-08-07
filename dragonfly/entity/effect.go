package entity

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/effect"
	"reflect"
	"sync"
	"time"
)

// EffectManager manages the effects of an entity. The effect manager will only store effects that last for
// a specific duration. Instant effects are applied instantly and not stored.
type EffectManager struct {
	mu      sync.Mutex
	effects map[reflect.Type]effect.Effect
}

// NewEffectManager creates and returns a new initialised EffectManager.
func NewEffectManager() *EffectManager {
	return &EffectManager{effects: map[reflect.Type]effect.Effect{}}
}

// Add adds an effect to the manager. If the effect is instant, it is applied to the Living entity passed
// immediately. If not, the effect is added to the EffectManager and is applied to the entity every time the
// Tick method is called.
// Effect levels of 0 or below will not do anything.
// Effect returns the final effect it added to the entity. That might be the effect passed or an effect with
// a higher level/duration than the one passed.
func (m *EffectManager) Add(e effect.Effect, entity Living) effect.Effect {
	if e.Level() <= 0 {
		return e
	}
	if e.Instant() {
		e.Apply(entity)
		return e
	}
	t := reflect.TypeOf(e)

	m.mu.Lock()
	existing, ok := m.effects[t]
	if !ok {
		m.effects[t] = e
		m.mu.Unlock()

		e.Start(entity)
		return e
	}
	if existing.Level() > e.Level() || (existing.Level() == e.Level() && existing.Duration() > e.Duration()) {
		m.mu.Unlock()
		return existing
	}
	m.effects[t] = e
	m.mu.Unlock()

	existing.End(entity)
	e.Start(entity)
	return e
}

// Remove removes any Effect present in the EffectManager with the type of the effect passed.
func (m *EffectManager) Remove(e effect.Effect, entity Living) {
	t := reflect.TypeOf(e)

	m.mu.Lock()
	existing, ok := m.effects[t]
	delete(m.effects, t)
	m.mu.Unlock()

	if ok {
		existing.End(entity)
	}
}

// Effects returns a list of all effects currently present in the effect manager. This will never include
// effects that have expired.
func (m *EffectManager) Effects() []effect.Effect {
	m.mu.Lock()
	defer m.mu.Unlock()

	e := make([]effect.Effect, 0, len(m.effects))
	for _, eff := range m.effects {
		e = append(e, eff)
	}
	return e
}

// Tick ticks the EffectManager, applying all of its effects to the Living entity passed when applicable and
// removing expired effects.
func (m *EffectManager) Tick(entity Living) {
	m.mu.Lock()
	e := make([]effect.Effect, 0, len(m.effects))
	var toEnd []effect.Effect

	for i, eff := range m.effects {
		e = append(e, eff)

		m.effects[i] = eff.WithSettings(eff.Duration()-time.Second/20, eff.Level(), eff.AmbientSource())
		if m.expired(eff) {
			delete(m.effects, i)
			toEnd = append(toEnd, eff)
		}
	}
	m.mu.Unlock()

	for _, eff := range e {
		eff.Apply(entity)
	}
	for _, eff := range toEnd {
		eff.End(entity)
	}
}

// expired checks if an Effect has expired.
func (m *EffectManager) expired(e effect.Effect) bool {
	return e.Duration() <= 0
}
