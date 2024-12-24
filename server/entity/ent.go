package entity

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"sync"
	"time"
)

// Behaviour implements the behaviour of an Ent.
type Behaviour interface {
	// Tick ticks the Ent using the Behaviour. A Movement is returned that
	// specifies the movement of the entity over the tick. Nil may be returned
	// if the entity did not move.
	Tick(e *Ent, tx *world.Tx) *Movement
}

// Ent is a world.Entity implementation that allows entity implementations to
// share a lot of code. It is currently under development and is prone to
// (breaking) changes.
type Ent struct {
	tx     *world.Tx
	handle *world.EntityHandle
	data   *world.EntityData
	once   sync.Once
}

// Open converts a world.EntityHandle to an Ent in a world.Tx.
func Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) *Ent {
	return &Ent{tx: tx, handle: handle, data: data}
}

func (e *Ent) H() *world.EntityHandle {
	return e.handle
}

func (e *Ent) Behaviour() Behaviour {
	return e.data.Data.(Behaviour)
}

// Explode propagates the explosion behaviour of the underlying Behaviour.
func (e *Ent) Explode(src mgl64.Vec3, impact float64, conf block.ExplosionConfig) {
	if expl, ok := e.Behaviour().(interface {
		Explode(e *Ent, src mgl64.Vec3, impact float64, conf block.ExplosionConfig)
	}); ok {
		expl.Explode(e, src, impact, conf)
	}
}

// Position returns the current position of the entity.
func (e *Ent) Position() mgl64.Vec3 {
	return e.data.Pos
}

// Velocity returns the current velocity of the entity. The values in the Vec3 returned represent the speed on
// that axis in blocks/tick.
func (e *Ent) Velocity() mgl64.Vec3 {
	return e.data.Vel
}

// SetVelocity sets the velocity of the entity. The values in the Vec3 passed represent the speed on
// that axis in blocks/tick.
func (e *Ent) SetVelocity(v mgl64.Vec3) {
	e.data.Vel = v
}

// Rotation returns the rotation of the entity.
func (e *Ent) Rotation() cube.Rotation {
	return e.data.Rot
}

// Age returns the total time lived of this entity. It increases by
// time.Second/20 for every time Tick is called.
func (e *Ent) Age() time.Duration {
	return e.data.Age
}

// OnFireDuration ...
func (e *Ent) OnFireDuration() time.Duration {
	return e.data.FireDuration
}

// SetOnFire ...
func (e *Ent) SetOnFire(duration time.Duration) {
	duration = max(duration, 0)
	stateChanged := (e.data.FireDuration > 0) != (duration > 0)

	e.data.FireDuration = duration
	if stateChanged {
		for _, v := range e.tx.Viewers(e.data.Pos) {
			v.ViewEntityState(e)
		}
	}
}

// Extinguish ...
func (e *Ent) Extinguish() {
	e.SetOnFire(0)
}

// NameTag returns the name tag of the entity. An empty string is returned if
// no name tag was set.
func (e *Ent) NameTag() string {
	return e.data.Name
}

// SetNameTag changes the name tag of an entity. The name tag is removed if an
// empty string is passed.
func (e *Ent) SetNameTag(s string) {
	e.data.Name = s
	for _, v := range e.tx.Viewers(e.Position()) {
		v.ViewEntityState(e)
	}
}

// Tick ticks Ent, progressing its lifetime and closing the entity if it is
// in the void.
func (e *Ent) Tick(tx *world.Tx, current int64) {
	y := e.data.Pos[1]
	if y < float64(tx.Range()[0]) && current%10 == 0 {
		_ = e.Close()
		return
	}
	e.SetOnFire(e.OnFireDuration() - time.Second/20)

	if m := e.Behaviour().Tick(e, tx); m != nil {
		m.Send()
	}
	e.data.Age += time.Second / 20
}

// Close closes the Ent and removes the associated entity from the world.
func (e *Ent) Close() error {
	e.once.Do(func() {
		e.tx.RemoveEntity(e)
		_ = e.handle.Close()
	})
	return nil
}
