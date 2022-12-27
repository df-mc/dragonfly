package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"sync"
	"time"
)

type Behaviour interface {
	Tick(e *Ent) *Movement
}

type Config struct {
	Behaviour Behaviour
}

func (conf Config) New(t world.EntityType, pos mgl64.Vec3) *Ent {
	return &Ent{t: t, pos: pos}
}

type Ent struct {
	conf Config
	t    world.EntityType

	mu  sync.Mutex
	pos mgl64.Vec3
	vel mgl64.Vec3
	rot cube.Rotation

	fireDuration time.Duration
}

func (e *Ent) Explode(src mgl64.Vec3, impact float64, conf block.ExplosionConfig) {
	if expl, ok := e.conf.Behaviour.(interface {
		Explode(e *Ent, src mgl64.Vec3, impact float64, conf block.ExplosionConfig)
	}); ok {
		expl.Explode(e, src, impact, conf)
	}
}

func NewEnt(t world.EntityType, pos mgl64.Vec3) *Ent {
	var conf Config
	// TODO: Default lifetime. What would the behaviour of that be?
	return conf.New(t, pos)
}

// Type returns the world.EntityType passed to NewEnt.
func (e *Ent) Type() world.EntityType {
	return e.t
}

// Owner returns the owner of the Ent, or nil if it doesn't have one.
func (e *Ent) Owner() world.Entity {
	// TODO: Change this signature to Owner() (world.Entity, bool) once all
	//  entities use this type.
	if owned, ok := e.conf.Behaviour.(interface {
		Owner() world.Entity
	}); ok {
		return owned.Owner()
	}
	return nil
}

// Position returns the current position of the entity.
func (e *Ent) Position() mgl64.Vec3 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.pos
}

// Velocity returns the current velocity of the entity. The values in the Vec3 returned represent the speed on
// that axis in blocks/tick.
func (e *Ent) Velocity() mgl64.Vec3 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.vel
}

// SetVelocity sets the velocity of the entity. The values in the Vec3 passed represent the speed on
// that axis in blocks/tick.
func (e *Ent) SetVelocity(v mgl64.Vec3) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vel = v
}

// Rotation always returns an empty cube.Rotation.
func (e *Ent) Rotation() cube.Rotation {
	return e.rot
}

// World returns the world of the entity.
func (e *Ent) World() *world.World {
	w, _ := world.OfEntity(e)
	return w
}

// OnFireDuration ...
func (e *Ent) OnFireDuration() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.fireDuration
}

// SetOnFire ...
func (e *Ent) SetOnFire(duration time.Duration) {
	if duration < 0 {
		duration = 0
	}
	e.mu.Lock()
	before, after := e.fireDuration > 0, duration > 0
	e.fireDuration = duration
	pos := e.pos
	e.mu.Unlock()

	if before == after {
		return
	}
	for _, v := range e.World().Viewers(pos) {
		v.ViewEntityState(e)
	}
}

// Extinguish ...
func (e *Ent) Extinguish() {
	e.SetOnFire(0)
}

// Tick ticks Ent, progressing its lifetime and closing the entity if it is
// in the void.
func (e *Ent) Tick(w *world.World, current int64) {
	if e.pos[1] < float64(w.Range()[0]) && current%10 == 0 {
		_ = e.Close()
		return
	}
	e.SetOnFire(e.OnFireDuration() - time.Second/20)

	if m := e.conf.Behaviour.Tick(e); m != nil {
		m.Send()
	}
}

// Close closes the Ent and removes the associated entity from the world.
func (e *Ent) Close() error {
	e.World().RemoveEntity(e)
	return nil
}
