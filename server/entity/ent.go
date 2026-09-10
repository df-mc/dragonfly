package entity

import (
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Behaviour implements the behaviour of an Ent.
type Behaviour interface {
	// Tick ticks the Ent using the Behaviour. A Movement is returned that
	// specifies the movement of the entity over the tick. Nil may be returned
	// if the entity did not move.
	Tick(e *Ent, tx *world.Tx) *Movement
}

// Ent is the live form of an entity within a single world.Tx: it is recreated through a world.EntityType's
// Open method for every transaction that touches the entity. It carries the base plumbing shared by all
// entity implementations. Persistent state must live in the entity's Behaviour, never in the Ent itself.
type Ent struct {
	tx                *world.Tx
	handle            *world.EntityHandle
	data              *world.EntityData
	deferPortalTravel bool
	once              sync.Once
}

// Open converts a world.EntityHandle to an Ent in a world.Tx.
func Open(tx *world.Tx, handle *world.EntityHandle, data *world.EntityData) *Ent {
	return &Ent{tx: tx, handle: handle, data: data}
}

// H returns the world.EntityHandle of the entity: its persistent form that outlives the transaction.
func (e *Ent) H() *world.EntityHandle {
	return e.handle
}

// Behaviour returns the Behaviour of the entity, stored in the entity data of its handle. Nil is returned
// if the entity data holds no Behaviour.
func (e *Ent) Behaviour() Behaviour {
	b, _ := e.data.Data.(Behaviour)
	return b
}

// Tx returns the transaction the entity was opened in.
func (e *Ent) Tx() *world.Tx {
	return e.tx
}

// Data returns the entity data of the entity's handle, the state that persists across transactions. It is
// writable, and writing to it is not the same as moving the entity: setting Pos or Vel here changes where the
// entity is without telling anyone watching it, so a viewer keeps drawing it where it was. Move an entity
// through its Movement instead, and reach for this only for state no method covers.
func (e *Ent) Data() *world.EntityData {
	return e.data
}

// Base returns the Ent itself. It is promoted by entities that embed Ent, so that the underlying Ent can be
// recognised wherever behaviour hooks are dispatched.
func (e *Ent) Base() *Ent {
	return e
}

// wrappedEnt is implemented by any entity carrying an Ent: the Ent itself or a struct embedding it.
type wrappedEnt interface {
	Base() *Ent
}

// Hurt dispatches damage to the entity's Behaviour if it implements DamageableBehaviour and reports the
// entity as invulnerable otherwise. It gives entities that are damageable without being alive the same Hurt
// method surface that code dealing damage asserts on Living entities.
func (e *Ent) Hurt(damage float64, src world.DamageSource) (n float64, vulnerable bool) {
	d, ok := e.Behaviour().(DamageableBehaviour)
	if !ok {
		return 0, false
	}
	ctx := e.tx.Event()
	if e.tx.World().Handler().HandleEntityHurt(ctx, e, &damage, src); ctx.Cancelled() {
		return 0, false
	}
	return d.Hurt(e, damage, src)
}

// ExplodableBehaviour may be implemented by a Behaviour to react to an explosion hitting its entity.
// Ent.Explode dispatches to it.
type ExplodableBehaviour interface {
	// Explode reacts to an explosion with the source and the impact on the entity passed.
	Explode(e *Ent, src world.ExplosionSource, impact float64)
}

// Explode propagates the explosion behaviour of the underlying Behaviour.
func (e *Ent) Explode(src world.ExplosionSource, impact float64) {
	if expl, ok := e.Behaviour().(ExplodableBehaviour); ok {
		expl.Explode(e, src, impact)
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

// Teleport teleports the entity to the position given.
func (e *Ent) Teleport(pos mgl64.Vec3) {
	viewers := e.tx.Viewers(e.data.Pos)
	e.data.Pos = pos
	for _, v := range viewers {
		v.ViewEntityTeleport(e, pos)
	}
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
		e.updateState()
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
	e.updateState()
}

// AlwaysShowNameTag returns whether the name tag of the entity is shown at all
// distances instead of only when the entity is looked at from up close.
func (e *Ent) AlwaysShowNameTag() bool {
	return e.data.AlwaysShowNameTag
}

// SetAlwaysShowNameTag changes whether the name tag of the entity is shown at
// all distances instead of only when the entity is looked at from up close.
func (e *Ent) SetAlwaysShowNameTag(alwaysShow bool) {
	e.data.AlwaysShowNameTag = alwaysShow
	e.updateState()
}

// UpdateState resends the entity's metadata to all viewers of the entity. Behaviours call it after changing
// state that is reflected in entity metadata.
func (e *Ent) UpdateState() {
	e.updateState()
}

// updateState updates the state of the entity for all viewers of the entity.
func (e *Ent) updateState() {
	for _, v := range e.tx.Viewers(e.data.Pos) {
		v.ViewEntityState(e)
	}
}

// PlayAction plays a world.EntityAction for all viewers of the entity.
func (e *Ent) PlayAction(a world.EntityAction) {
	for _, v := range e.tx.Viewers(e.data.Pos) {
		v.ViewEntityAction(e, a)
	}
}

// Tick ticks Ent, progressing its lifetime and closing the entity if it is
// in the void.
func (e *Ent) Tick(tx *world.Tx, current int64) {
	e.deferPortalTravel = true
	defer func() {
		e.deferPortalTravel = false
	}()

	y := e.data.Pos[1]
	if y < float64(tx.Range()[0]) && current%10 == 0 {
		// Living entities are hurt by the void in LivingEnt.Tick instead of vanishing silently.
		if _, living := e.Behaviour().(LivingBehaviour); !living {
			_ = e.Close()
			return
		}
	}
	e.SetOnFire(e.OnFireDuration() - time.Second/20)

	m := e.Behaviour().Tick(e, tx)
	if e.finishPendingPortalTravel(tx) {
		return
	}
	if m != nil {
		m.Send()
	}
	if e.checkPortalInsiders() && e.finishPendingPortalTravel(tx) {
		return
	}
	e.stopPortalContact()
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

// TravelThroughPortal handles the entity touching a portal block.
func (e *Ent) TravelThroughPortal(tx *world.Tx, target world.Dimension) {
	if tc := e.portalTravelComputer(); tc != nil {
		if e.deferPortalTravel {
			tc.queuePortalTravel(tx, target)
			return
		}
		tc.EnterPortal(e, tx, target)
	}
}

// portalTravelComputer returns the behaviour's portal travel state, if any.
func (e *Ent) portalTravelComputer() *PortalTravelComputer {
	if b, ok := e.Behaviour().(PortalTravelComputerProvider); ok {
		return b.PortalTravelComputer()
	}
	return nil
}

// stopPortalContact resets portal contact state when no portal was touched.
func (e *Ent) stopPortalContact() {
	if tc := e.portalTravelComputer(); tc != nil {
		tc.StopPortalContact()
	}
}

// pendingPortalTravel reports whether this tick queued terminal portal travel.
func (e *Ent) pendingPortalTravel() bool {
	if tc := e.portalTravelComputer(); tc != nil {
		return tc.hasPendingPortalTravel()
	}
	return false
}

// finishPendingPortalTravel starts queued terminal portal travel, if present.
func (e *Ent) finishPendingPortalTravel(tx *world.Tx) bool {
	if tc := e.portalTravelComputer(); tc != nil {
		return tc.finishPendingPortalTravel(e, tx)
	}
	return false
}

type portalBlock interface {
	Portal() world.Dimension
}

// checkPortalInsiders checks whether the entity is inside portal blocks.
// Other EntityInsider blocks are intentionally left to entity physics.
func (e *Ent) checkPortalInsiders() bool {
	box := e.H().Type().BBox(e).Translate(e.Position()).Grow(-0.0001)
	low, high := cube.PosFromVec3(box.Min()), cube.PosFromVec3(box.Max())

	for blockPos := range cube.Range3D(low, high) {
		if p, ok := e.tx.Block(blockPos).(portalBlock); ok {
			e.TravelThroughPortal(e.tx, p.Portal())
			if e.pendingPortalTravel() {
				return true
			}
		}
	}
	return false
}

// boolByte returns 1 if the bool passed is true, or 0 if it is false.
func boolByte(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
