package world

import (
	"iter"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world/redstone"
	"github.com/go-gl/mathgl/mgl64"
)

// Tx is a transitional alias for Context that keeps existing *world.Tx
// signatures compiling. New code should use Context; a follow-up removes it.
type Tx = Context

// Context is the owner-scoped handle passed to world callbacks; it is the only
// way to perform world operations and is valid only during its callback.
// Off-owner code obtains one via World.Do, world.Call, EntityRef.Do, or player.Ref.
type Context struct {
	*tx

	// cancel records whether an event Context was cancelled by a handler.
	cancel bool
}

// tx is the unexported transaction state shared by every Context derived from
// it (via Event) within one owner callback.
type tx struct {
	w        *World
	closed   bool
	deferred []scheduledTransaction
}

// contextAlloc bundles a Context with its transaction so newContext costs one
// allocation on the per-transaction hot path.
type contextAlloc struct {
	ctx Context
	tx  tx
}

// newContext returns a Context backed by a fresh transaction on World w.
func newContext(w *World) *Context {
	a := &contextAlloc{tx: tx{w: w}}
	a.ctx.tx = &a.tx
	return &a.ctx
}

// Event returns a Context sharing ctx's transaction but with its own cancel
// state, so dispatching one Handler event can't cancel another in the same
// transaction. Only code that fires world/player Handler events needs it.
func (ctx *Context) Event() *Context {
	return &Context{tx: ctx.tx}
}

// Cancelled returns whether the Context has been cancelled by an event handler.
func (ctx *Context) Cancelled() bool { return ctx.cancel }

// Cancel cancels the Context. It is used by event handlers to signal that the
// default behaviour of the event should not run.
func (ctx *Context) Cancel() { ctx.cancel = true }

// Defer schedules f to run on the owner after the current callback completes.
func (ctx *Context) Defer(f func(ctx *Context)) *Task {
	return ctx.DeferErr(func(ctx *Context) error {
		f(ctx)
		return nil
	})
}

// DeferErr schedules f to run on the owner after the current callback
// completes, recording any returned error on the Task.
func (ctx *Context) DeferErr(f func(ctx *Context) error) *Task {
	return ctx.deferTask(f)
}

// Range returns the lower and upper bounds of the World that the Context is
// operating on.
func (ctx *Context) Range() cube.Range {
	return ctx.World().ra
}

// SetBlock writes a block to the position passed. If a chunk is not yet loaded
// at that position, the chunk is first loaded or generated if it could not be
// found in the world save. SetBlock panics if the block passed has not yet
// been registered using RegisterBlock(). Nil may be passed as the block to set
// the block to air.
//
// A SetOpts struct may be passed to additionally modify behaviour of SetBlock,
// specifically to improve performance under specific circumstances. Nil should
// be passed where performance is not essential, to make sure the world is
// updated adequately.
//
// SetBlock should be avoided in situations where performance is critical when
// needing to set a lot of blocks to the world. BuildStructure may be used
// instead.
func (ctx *Context) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {
	ctx.World().setBlock(pos, b, opts)
}

// Block reads a block from the position passed. If a chunk is not yet loaded
// at that position, the chunk is loaded, or generated if it could not be found
// in the world save, and the block returned.
func (ctx *Context) Block(pos cube.Pos) Block {
	return ctx.World().block(pos)
}

// Liquid attempts to return a Liquid block at the position passed. This
// Liquid may be in the foreground or in any other layer. If found, the Liquid
// is returned. If not, the bool returned is false.
func (ctx *Context) Liquid(pos cube.Pos) (Liquid, bool) {
	return ctx.World().liquid(pos)
}

// SetLiquid sets a Liquid at a specific position in the World. Unlike
// SetBlock, SetLiquid will not necessarily overwrite any existing blocks. It
// will instead be in the same position as a block currently there, unless
// there already is a Liquid at that position, in which case it will be
// overwritten. If nil is passed for the Liquid, any Liquid currently present
// will be removed.
func (ctx *Context) SetLiquid(pos cube.Pos, b Liquid) {
	ctx.World().setLiquid(pos, b)
}

// BuildStructure builds a Structure passed at a specific position in the
// world. Unlike SetBlock, it takes a Structure implementation, which provides
// blocks to be placed at a specific location. BuildStructure is specifically
// optimised to be able to process a large batch of chunks simultaneously and
// will do so within much less time than separate SetBlock calls would. The
// method operates on a per-chunk basis, setting all blocks within a single
// chunk part of the Structure before moving on to the next chunk.
func (ctx *Context) BuildStructure(pos cube.Pos, s Structure) {
	ctx.World().buildStructure(pos, s)
}

// ScheduleBlockUpdate schedules a block update at the position passed for the
// block type passed after a specific delay. If the block at that position does
// not handle block updates, nothing will happen.
// Block updates are both block and position specific. A block update is only
// scheduled if no block update with the same position and block type is
// already scheduled at a later time than the newly scheduled update.
func (ctx *Context) ScheduleBlockUpdate(pos cube.Pos, b Block, delay time.Duration) {
	ctx.World().scheduleBlockUpdate(pos, b, delay)
}

// HighestLightBlocker gets the Y value of the highest fully light blocking
// block at the x and z values passed in the World.
func (ctx *Context) HighestLightBlocker(x, z int) int {
	return ctx.World().HighestLightBlocker(x, z)
}

// HighestBlock looks up the highest non-air block in the World at a specific x
// and z. The y value of the highest block is returned, or 0 if no blocks were
// present in the column.
func (ctx *Context) HighestBlock(x, z int) int {
	return ctx.World().highestBlock(x, z)
}

// Light returns the light level at the position passed. This is the highest of
// the sky- and block light. The light value returned is a value in the range
// 0-15, where 0 means there is no light present, whereas 15 means the block is
// fully lit.
func (ctx *Context) Light(pos cube.Pos) uint8 {
	return ctx.World().light(pos)
}

// SkyLight returns the skylight level at the position passed. This light level
// is not influenced by blocks that emit light, such as torches. The light
// value, similarly to Light, is a value in the range 0-15, where 0 means no
// light is present.
func (ctx *Context) SkyLight(pos cube.Pos) uint8 {
	return ctx.World().skyLight(pos)
}

// SetBiome sets the Biome at the position passed. If a chunk is not yet loaded
// at that position, the chunk is first loaded or generated if it could not be
// found in the world save.
func (ctx *Context) SetBiome(pos cube.Pos, b Biome) {
	ctx.World().setBiome(pos, b)
}

// Biome reads the Biome at the position passed. If a chunk is not yet loaded
// at that position, the chunk is loaded, or generated if it could not be found
// in the world save, and the Biome returned.
func (ctx *Context) Biome(pos cube.Pos) Biome {
	return ctx.World().biome(pos)
}

// Temperature returns the temperature in the World at a specific position.
// Higher altitudes and different biomes influence the temperature returned.
func (ctx *Context) Temperature(pos cube.Pos) float64 {
	return ctx.World().temperature(pos)
}

// RainingAt checks if it is raining at a specific cube.Pos in the World. True
// is returned if it is raining, if the temperature is high enough in the biome
// for it not to be snow and if the block is above the top-most obstructing
// block.
func (ctx *Context) RainingAt(pos cube.Pos) bool {
	return ctx.World().rainingAt(pos)
}

// SnowingAt checks if it is snowing at a specific cube.Pos in the World. True
// is returned if the temperature in the Biome at that position is sufficiently
// low, if it is raining and if it's above the top-most obstructing block.
func (ctx *Context) SnowingAt(pos cube.Pos) bool {
	return ctx.World().snowingAt(pos)
}

// ThunderingAt checks if it is thundering at a specific cube.Pos in the World.
// True is returned if RainingAt returns true and if it is thundering in the
// world.
func (ctx *Context) ThunderingAt(pos cube.Pos) bool {
	return ctx.World().thunderingAt(pos)
}

// Raining checks if it is raining anywhere in the World.
func (ctx *Context) Raining() bool {
	return ctx.World().raining()
}

// Thundering checks if it is thundering anywhere in the World.
func (ctx *Context) Thundering() bool {
	return ctx.World().thundering()
}

// AddParticle spawns a Particle at a given position in the World. Viewers that
// are viewing the chunk will be shown the particle.
func (ctx *Context) AddParticle(pos mgl64.Vec3, p Particle) {
	ctx.World().addParticle(pos, p)
}

// PlayEntityAnimation plays an animation on an entity in the World. The animation is played for all viewers
// of the entity.
func (ctx *Context) PlayEntityAnimation(e Entity, a EntityAnimation) {
	for _, viewer := range ctx.World().viewersOf(e.Position()) {
		viewer.ViewEntityAnimation(e, a)
	}
}

// PlaySound plays a sound at a specific position in the World. Viewers of that
// position will be able to hear the sound if they are close enough.
func (ctx *Context) PlaySound(pos mgl64.Vec3, s Sound) {
	ctx.World().playSound(ctx, pos, s)
}

// AddEntity adds an EntityHandle to a World. The Entity will be visible to all
// viewers of the World that have the chunk at the EntityHandle's position. If
// the chunk that the EntityHandle is in is not yet loaded, it will first be
// loaded. AddEntity panics if the EntityHandle is already in a world.
// AddEntity returns the Entity created by the EntityHandle.
func (ctx *Context) AddEntity(e *EntityHandle) Entity {
	return ctx.World().addEntity(ctx, e)
}

// RemoveEntity removes an Entity from the World that is currently present in
// it. Any viewers of the Entity will no longer be able to see it.
// RemoveEntity returns the EntityHandle of the Entity. After removing an Entity
// from the World, the Entity is no longer usable.
func (ctx *Context) RemoveEntity(e Entity) *EntityHandle {
	return ctx.World().removeEntity(e, ctx)
}

// EntitiesWithin returns an iterator that yields all entities contained within
// the cube.BBox passed.
func (ctx *Context) EntitiesWithin(box cube.BBox) iter.Seq[Entity] {
	return ctx.World().entitiesWithin(ctx, box)
}

// Entities returns an iterator that yields all entities in the World.
func (ctx *Context) Entities() iter.Seq[Entity] {
	return ctx.World().allEntities(ctx)
}

// Players returns an iterator that yields all player entities in the World.
func (ctx *Context) Players() iter.Seq[Entity] {
	return ctx.World().allPlayers(ctx)
}

// Viewers returns all viewers viewing the position passed.
func (ctx *Context) Viewers(pos mgl64.Vec3) []Viewer {
	return ctx.World().viewersOf(pos)
}

// Sleepers returns an iterator that yields all sleeping entities currently added to the World.
func (ctx *Context) Sleepers() iter.Seq[Sleeper] {
	ent := ctx.Entities()
	return func(yield func(Sleeper) bool) {
		for e := range ent {
			if sleeper, ok := e.(Sleeper); ok {
				if !yield(sleeper) {
					return
				}
			}
		}
	}
}

// BroadcastSleepingIndicator broadcasts a sleeping indicator to all sleepers in the world.
func (ctx *Context) BroadcastSleepingIndicator() {
	sleepers := ctx.Sleepers()

	var sleeping, allSleepers int
	for s := range sleepers {
		allSleepers++
		if _, ok := s.Sleeping(); ok {
			sleeping++
		}
	}

	for s := range sleepers {
		s.SendSleepingIndicator(sleeping, allSleepers)
	}
}

// BroadcastSleepingReminder broadcasts a sleeping reminder message to all sleepers in the world, excluding the sleeper
// passed.
func (ctx *Context) BroadcastSleepingReminder(sleeper Sleeper) {
	sleepers := ctx.Sleepers()

	var notSleeping int
	for s := range sleepers {
		if _, ok := s.Sleeping(); !ok {
			notSleeping++
		}
	}

	for s := range sleepers {
		if _, ok := s.Sleeping(); !ok {
			s.Messaget(chat.MessageSleeping, sleeper.Name(), notSleeping)
		}
	}
}

// RedstonePower returns the redstone power emitted by the block at pos toward a neighbouring receiver.
// The face argument is relative to the receiving block.
func (ctx *Context) RedstonePower(pos cube.Pos, face cube.Face, accountForDust bool) (power int) {
	b := ctx.Block(pos)
	if c, ok := b.(Conductor); ok {
		return c.WeakPower(pos, face, ctx, accountForDust)
	}
	// The wiki states that in the future some blocks may be transparent but still relay redstone.
	// If a block implements RedstonePowerRelayer, it should always be prioritised over lightDiffuser.
	if r, ok := b.(RedstonePowerRelayer); ok {
		if !r.RelaysRedstonePowerThrough() {
			return 0
		}
	} else if d, ok := b.(lightDiffuser); ok && d.LightDiffusionLevel() != 15 {
		return 0
	}
	for _, f := range cube.Faces() {
		if !b.Model().FaceSolid(pos, f, ctx) {
			return 0
		}
	}
	for _, f := range cube.Faces() {
		c, ok := ctx.Block(pos.Side(f)).(Conductor)
		if !ok {
			continue
		}
		sourcePos := pos.Side(f)
		power = max(power, c.StrongPower(sourcePos, f, ctx, accountForDust))
		if !accountForDust {
			continue
		}
		if weakBlockPowerer, ok := c.(WeakBlockPowerer); ok && weakBlockPowerer.WeaklyPowersBlocks() {
			power = max(power, c.WeakPower(sourcePos, f, ctx, accountForDust))
		}
	}
	return power
}

func (ctx *Context) deferTask(f func(ctx *Context) error) *Task {
	task := newTask()
	if ctx.closed {
		task.failIfPending(ErrWorldClosed)
		return task
	}
	ctx.deferred = append(ctx.deferred, scheduledTransaction{task: task, f: f})
	return task
}

// World returns the Context's World. It panics once the callback has
// completed. Treat the result as the off-owner handle: blocking calls like
// Save and Close deadlock from inside the callback, so do world operations
// through the Context instead.
func (ctx *Context) World() *World {
	if ctx.closed {
		panic("world.Context: use of transaction after transaction finishes is not permitted")
	}
	return ctx.w
}

// CurrentTick returns the current tick of the transaction's world.
func (ctx *Context) CurrentTick() int64 {
	w := ctx.World()
	w.set.Lock()
	defer w.set.Unlock()
	return w.set.CurrentTick
}

// Redstone returns the transient redstone runtime state owned by the transaction's world.
func (ctx *Context) Redstone() *redstone.State {
	return &ctx.World().redstone
}

// close finishes the Context, causing any following call on the Context to panic.
func (ctx *Context) close() {
	ctx.closed = true
}

func (ctx *Context) runDeferred() {
	for len(ctx.deferred) > 0 {
		deferred := ctx.deferred
		ctx.deferred = nil
		for _, st := range deferred {
			st.Run(ctx.w)
		}
	}
}

// normalTransaction is added to the transaction queue for transactions created
// using World.exec().
type normalTransaction struct {
	c chan struct{}
	f func(ctx *Context)
}

// Run creates a *Context, calls ntx.f, closes the transaction and finally closes
// ntx.c.
func (ntx normalTransaction) Run(w *World) {
	ctx := newContext(w)
	ntx.f(ctx)
	ctx.close()
	ctx.runDeferred()
	close(ntx.c)
}

// weakTransaction is a transaction that may be cancelled by its validity
// predicate before the transaction is run.
type weakTransaction struct {
	c     chan bool
	f     func(ctx *Context)
	valid func() bool
	cond  *sync.Cond
}

// Run runs the transaction, first checking if it is still valid and creating a
// *Context if so. Afterwards, a bool indicating if the transaction was run is added
// to wtx.c. Finally, wtx.cond.Broadcast() is called.
func (wtx weakTransaction) Run(w *World) {
	valid := wtx.valid == nil || wtx.valid()
	if valid {
		ctx := newContext(w)
		wtx.f(ctx)
		ctx.close()
		ctx.runDeferred()
	}
	// We have to acquire a lock on wtx.cond.L here to make sure cond.Wait()
	// has been called before we call cond.Broadcast(). If not, we might
	// broadcast before cond.Wait() and cause a permanent suspension.
	wtx.cond.L.Lock()
	defer wtx.cond.L.Unlock()

	wtx.c <- valid
	wtx.cond.Broadcast()
}

// fail delivers false to a weak transaction that will never run, using the
// same condition handshake as Run so a waiter in cond.Wait is woken.
func (wtx weakTransaction) fail() {
	wtx.cond.L.Lock()
	defer wtx.cond.L.Unlock()
	wtx.c <- false
	wtx.cond.Broadcast()
}
