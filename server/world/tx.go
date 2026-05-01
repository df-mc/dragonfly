package world

import (
	"iter"
	"sync"
	"sync/atomic"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/go-gl/mathgl/mgl64"
)

// Tx represents a synchronised transaction performed on a World. Most
// operations on a World can only be called through a transaction. Tx is not
// safe for use by multiple goroutines concurrently.
type Tx struct {
	w      *World
	closed bool
}

// Range returns the lower and upper bounds of the World that the Tx is
// operating on.
func (tx *Tx) Range() cube.Range {
	return tx.w.ra
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
func (tx *Tx) SetBlock(pos cube.Pos, b Block, opts *SetOpts) {
	tx.World().setBlock(pos, b, opts)
}

// Block reads a block from the position passed. If a chunk is not yet loaded
// at that position, the chunk is loaded, or generated if it could not be found
// in the world save, and the block returned.
func (tx *Tx) Block(pos cube.Pos) Block {
	return tx.World().block(pos)
}

// Liquid attempts to return a Liquid block at the position passed. This
// Liquid may be in the foreground or in any other layer. If found, the Liquid
// is returned. If not, the bool returned is false.
func (tx *Tx) Liquid(pos cube.Pos) (Liquid, bool) {
	return tx.World().liquid(pos)
}

// SetLiquid sets a Liquid at a specific position in the World. Unlike
// SetBlock, SetLiquid will not necessarily overwrite any existing blocks. It
// will instead be in the same position as a block currently there, unless
// there already is a Liquid at that position, in which case it will be
// overwritten. If nil is passed for the Liquid, any Liquid currently present
// will be removed.
func (tx *Tx) SetLiquid(pos cube.Pos, b Liquid) {
	tx.World().setLiquid(pos, b)
}

// BuildStructure builds a Structure passed at a specific position in the
// world. Unlike SetBlock, it takes a Structure implementation, which provides
// blocks to be placed at a specific location. BuildStructure is specifically
// optimised to be able to process a large batch of chunks simultaneously and
// will do so within much less time than separate SetBlock calls would. The
// method operates on a per-chunk basis, setting all blocks within a single
// chunk part of the Structure before moving on to the next chunk.
func (tx *Tx) BuildStructure(pos cube.Pos, s Structure) {
	tx.World().buildStructure(pos, s)
}

// ScheduleBlockUpdate schedules a block update at the position passed for the
// block type passed after a specific delay. If the block at that position does
// not handle block updates, nothing will happen.
// Block updates are both block and position specific. A block update is only
// scheduled if no block update with the same position and block type is
// already scheduled at a later time than the newly scheduled update.
func (tx *Tx) ScheduleBlockUpdate(pos cube.Pos, b Block, delay time.Duration) {
	tx.World().scheduleBlockUpdate(pos, b, delay)
}

// HighestLightBlocker gets the Y value of the highest fully light blocking
// block at the x and z values passed in the World.
func (tx *Tx) HighestLightBlocker(x, z int) int {
	return tx.World().highestLightBlocker(x, z)
}

// HighestBlock looks up the highest non-air block in the World at a specific x
// and z. The y value of the highest block is returned, or 0 if no blocks were
// present in the column.
func (tx *Tx) HighestBlock(x, z int) int {
	return tx.World().highestBlock(x, z)
}

// Light returns the light level at the position passed. This is the highest of
// the sky- and block light. The light value returned is a value in the range
// 0-15, where 0 means there is no light present, whereas 15 means the block is
// fully lit.
func (tx *Tx) Light(pos cube.Pos) uint8 {
	return tx.World().light(pos)
}

// SkyLight returns the skylight level at the position passed. This light level
// is not influenced by blocks that emit light, such as torches. The light
// value, similarly to Light, is a value in the range 0-15, where 0 means no
// light is present.
func (tx *Tx) SkyLight(pos cube.Pos) uint8 {
	return tx.World().skyLight(pos)
}

// SetBiome sets the Biome at the position passed. If a chunk is not yet loaded
// at that position, the chunk is first loaded or generated if it could not be
// found in the world save.
func (tx *Tx) SetBiome(pos cube.Pos, b Biome) {
	tx.World().setBiome(pos, b)
}

// Biome reads the Biome at the position passed. If a chunk is not yet loaded
// at that position, the chunk is loaded, or generated if it could not be found
// in the world save, and the Biome returned.
func (tx *Tx) Biome(pos cube.Pos) Biome {
	return tx.World().biome(pos)
}

// Temperature returns the temperature in the World at a specific position.
// Higher altitudes and different biomes influence the temperature returned.
func (tx *Tx) Temperature(pos cube.Pos) float64 {
	return tx.World().temperature(pos)
}

// RainingAt checks if it is raining at a specific cube.Pos in the World. True
// is returned if it is raining, if the temperature is high enough in the biome
// for it not to be snow and if the block is above the top-most obstructing
// block.
func (tx *Tx) RainingAt(pos cube.Pos) bool {
	return tx.World().rainingAt(pos)
}

// SnowingAt checks if it is snowing at a specific cube.Pos in the World. True
// is returned if the temperature in the Biome at that position is sufficiently
// low, if it is raining and if it's above the top-most obstructing block.
func (tx *Tx) SnowingAt(pos cube.Pos) bool {
	return tx.World().snowingAt(pos)
}

// ThunderingAt checks if it is thundering at a specific cube.Pos in the World.
// True is returned if RainingAt returns true and if it is thundering in the
// world.
func (tx *Tx) ThunderingAt(pos cube.Pos) bool {
	return tx.World().thunderingAt(pos)
}

// Raining checks if it is raining anywhere in the World.
func (tx *Tx) Raining() bool {
	return tx.World().raining()
}

// Thundering checks if it is thundering anywhere in the World.
func (tx *Tx) Thundering() bool {
	return tx.World().thundering()
}

// AddParticle spawns a Particle at a given position in the World. Viewers that
// are viewing the chunk will be shown the particle.
func (tx *Tx) AddParticle(pos mgl64.Vec3, p Particle) {
	tx.World().addParticle(pos, p)
}

// PlayEntityAnimation plays an animation on an entity in the World. The animation is played for all viewers
// of the entity.
func (tx *Tx) PlayEntityAnimation(e Entity, a EntityAnimation) {
	for _, viewer := range tx.World().viewersOf(e.Position()) {
		viewer.ViewEntityAnimation(e, a)
	}
}

// PlaySound plays a sound at a specific position in the World. Viewers of that
// position will be able to hear the sound if they are close enough.
func (tx *Tx) PlaySound(pos mgl64.Vec3, s Sound) {
	tx.World().playSound(tx, pos, s)
}

// AddEntity adds an EntityHandle to a World. The Entity will be visible to all
// viewers of the World that have the chunk at the EntityHandle's position. If
// the chunk that the EntityHandle is in is not yet loaded, it will first be
// loaded. AddEntity panics if the EntityHandle is already in a world.
// AddEntity returns the Entity created by the EntityHandle.
func (tx *Tx) AddEntity(e *EntityHandle) Entity {
	return tx.World().addEntity(tx, e)
}

// RemoveEntity removes an Entity from the World that is currently present in
// it. Any viewers of the Entity will no longer be able to see it.
// RemoveEntity returns the EntityHandle of the Entity. After removing an Entity
// from the World, the Entity is no longer usable.
func (tx *Tx) RemoveEntity(e Entity) *EntityHandle {
	return tx.World().removeEntity(e, tx)
}

// EntitiesWithin returns an iterator that yields all entities contained within
// the cube.BBox passed.
func (tx *Tx) EntitiesWithin(box cube.BBox) iter.Seq[Entity] {
	return tx.World().entitiesWithin(tx, box)
}

// Entities returns an iterator that yields all entities in the World.
func (tx *Tx) Entities() iter.Seq[Entity] {
	return tx.World().allEntities(tx)
}

// Players returns an iterator that yields all player entities in the World.
func (tx *Tx) Players() iter.Seq[Entity] {
	return tx.World().allPlayers(tx)
}

// Viewers returns all viewers viewing the position passed.
func (tx *Tx) Viewers(pos mgl64.Vec3) []Viewer {
	return tx.World().viewersOf(pos)
}

// Sleepers returns an iterator that yields all sleeping entities currently added to the World.
func (tx *Tx) Sleepers() iter.Seq[Sleeper] {
	ent := tx.Entities()
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
func (tx *Tx) BroadcastSleepingIndicator() {
	sleepers := tx.Sleepers()

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
func (tx *Tx) BroadcastSleepingReminder(sleeper Sleeper) {
	sleepers := tx.Sleepers()

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

// World returns the World of the Tx. It panics if the transaction was already
// marked complete.
func (tx *Tx) World() *World {
	if tx.closed {
		panic("world.Tx: use of transaction after transaction finishes is not permitted")
	}
	return tx.w
}

// close finishes the Tx, causing any following call on the Tx to panic.
func (tx *Tx) close() {
	tx.closed = true
}

// normalTransaction is added to the transaction queue for transactions created
// using World.Exec().
type normalTransaction struct {
	c chan struct{}
	f func(tx *Tx)
}

// Run creates a *Tx, calls ntx.f, closes the transaction and finally closes
// ntx.c.
func (ntx normalTransaction) Run(w *World) {
	tx := &Tx{w: w}
	ntx.f(tx)
	tx.close()
	close(ntx.c)
}

// weakTransaction is a transaction that may be cancelled by setting its invalid
// bool to false before the transaction is run.
type weakTransaction struct {
	c       chan bool
	f       func(tx *Tx)
	invalid *atomic.Bool
	cond    *sync.Cond
}

// Run runs the transaction, first checking if its invalid bool is false and
// creating a *Tx if so. Afterwards, a bool indicating if the transaction was
// run is added to wtx.c. Finally, wtx.cond.Broadcast() is called.
func (wtx weakTransaction) Run(w *World) {
	valid := !wtx.invalid.Load()
	if valid {
		tx := &Tx{w: w}
		wtx.f(tx)
		tx.close()
	}
	// We have to acquire a lock on wtx.cond.L here to make sure cond.Wait()
	// has been called before we call cond.Broadcast(). If not, we might
	// broadcast before cond.Wait() and cause a permanent suspension.
	wtx.cond.L.Lock()
	defer wtx.cond.L.Unlock()

	wtx.c <- valid
	wtx.cond.Broadcast()
}
