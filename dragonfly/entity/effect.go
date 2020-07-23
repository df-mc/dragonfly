package entity

import (
	"image/color"
	"reflect"
	"sync"
	"time"
)

// Effect represents an effect that may be added to a living entity. Effects may either be instant or last
// for a specific duration.
type Effect interface {
	// Instant checks if the effect is instance. If it is instant, the effect will only be ticked a single
	// time when added to an entity.
	Instant() bool
	// Apply applies the effect to an entity. For instant effects, this method applies the effect once, such
	// as healing the Living entity for instant health.
	Apply(e Living)
	// Level returns the level of the effect. A higher level generally means a more powerful effect.
	Level() int
	// Duration returns the leftover duration of the effect.
	Duration() time.Duration
	// WithDurationAndLevel returns the effect with a duration and level passed.
	WithDurationAndLevel(d time.Duration, level int) Effect
	// RGBA returns the colour of the effect. If multiple effects are present, the colours will be mixed
	// together to form a new colour.
	RGBA() color.RGBA
	// ShowParticles checks if the particle should show particles. If not, entities that have the effect
	// will not display particles around them.
	ShowParticles() bool
	// AmbientSource specifies if the effect came from an ambient source, such as a beacon or conduit. The
	// particles will be less visible when this is true.
	AmbientSource() bool
	// Start is called for lasting events. It is sent the first time the effect is applied to an entity.
	Start(e Living)
	// End is called for lasting events. It is sent the moment the effect expires.
	End(e Living)
}

// EffectManager manages the effects of an entity. The effect manager will only store effects that last for
// a specific duration. Instant effects are applied instantly and not stored.
type EffectManager struct {
	mu      sync.Mutex
	effects map[reflect.Type]Effect
}

// NewEffectManager creates and returns a new initialised EffectManager.
func NewEffectManager() *EffectManager {
	return &EffectManager{effects: map[reflect.Type]Effect{}}
}

// Add adds an effect to the manager. If the effect is instant, it is applied to the Living entity passed
// immediately. If not, the effect is added to the EffectManager and is applied to the entity every time the
// Tick method is called.
// Effect levels of 0 or below will not do anything.
func (m *EffectManager) Add(e Effect, entity Living) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e.Level() <= 0 {
		return
	}

	if e.Instant() {
		e.Apply(entity)
		return
	}
	t := reflect.TypeOf(e)
	existing, ok := m.effects[t]
	if !ok {
		m.effects[t] = e
		e.Start(entity)
		return
	}
	if existing.Level() > e.Level() || (existing.Level() == e.Level() && existing.Duration() > e.Duration()) {
		return
	}
	existing.End(entity)
	m.effects[t] = e
	e.Start(entity)
}

// Remove removes any Effect present in the EffectManager with the type of the effect passed.
func (m *EffectManager) Remove(e Effect, entity Living) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t := reflect.TypeOf(e)
	if existing, ok := m.effects[t]; ok {
		existing.End(entity)
	}
	delete(m.effects, t)
}

// Effects returns a list of all effects currently present in the effect manager. This will never include
// effects that have expired.
func (m *EffectManager) Effects() []Effect {
	m.mu.Lock()
	defer m.mu.Unlock()

	e := make([]Effect, 0, len(m.effects))
	for _, effect := range m.effects {
		e = append(e, effect)
	}
	return e
}

// Tick ticks the EffectManager, applying all of its effects to the Living entity passed when applicable and
// removing expired effects.
func (m *EffectManager) Tick(entity Living) {
	m.mu.Lock()
	e := make([]Effect, 0, len(m.effects))
	for i, effect := range m.effects {
		e = append(e, effect)

		m.effects[i] = effect.WithDurationAndLevel(effect.Duration()-time.Second/20, effect.Level())
		if m.expired(effect) {
			delete(m.effects, i)
			effect.End(entity)
			continue
		}
	}
	m.mu.Unlock()

	for _, effect := range e {
		effect.Apply(entity)
	}
}

// expired checks if an Effect has expired.
func (m *EffectManager) expired(e Effect) bool {
	return e.Duration() <= 0
}
