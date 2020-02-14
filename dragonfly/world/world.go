package world

import (
	"context"
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/chunk"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/gamemode"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/particle"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"
)

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
// World generally provides a synchronised state: All entities, blocks and players usually operate in this
// world, so World ensures that all its methods will always be safe for simultaneous calls.
type World struct {
	name string
	log  *logrus.Logger

	stopTick   context.Context
	cancelTick context.CancelFunc

	time        int64
	timeStopped uint32

	gamemodeMu      sync.RWMutex
	defaultGameMode gamemode.GameMode

	hMutex sync.RWMutex
	hand   Handler

	pMutex sync.RWMutex
	prov   Provider
	cCache *cache.Cache

	gMutex sync.RWMutex
	gen    Generator

	entityMutex sync.RWMutex
	entities    map[ChunkPos][]Entity

	viewerMutex sync.RWMutex
	viewers     map[ChunkPos][]Viewer
}

// New creates a new initialised world. The world may be used right away, but it will not be saved or loaded
// from files until it has been given a different provider than the default. (NoIOProvider)
// By default, the name of the world will be 'World'.
func New(log *logrus.Logger) *World {
	ctx, cancel := context.WithCancel(context.Background())
	w := &World{
		name:            "World",
		prov:            NoIOProvider{},
		gen:             FlatGenerator{},
		log:             log,
		viewers:         make(map[ChunkPos][]Viewer),
		entities:        make(map[ChunkPos][]Entity),
		stopTick:        ctx,
		cancelTick:      cancel,
		defaultGameMode: gamemode.Survival{},
	}
	w.initChunkCache()
	go w.startTicking()
	return w
}

// Name returns the display name of the world. Generally, this name is displayed at the top of the player list
// in the pause screen in-game.
// If a provider is set, the name will be updated according to the name that it provides.
func (w *World) Name() string {
	w.pMutex.RLock()
	n := w.name
	w.pMutex.RUnlock()
	return n
}

// Block reads a block from the position passed. If a chunk is not yet loaded at that position, the chunk is
// loaded, or generated if it could not be found in the world save, and the block returned. Chunks will be
// loaded synchronously.
// An error is returned if the chunk that the block is located in could not be loaded successfully.
func (w *World) Block(pos block.Position) (block.Block, error) {
	c, err := w.chunk(chunkPosFromBlockPos(pos), true)
	if err != nil {
		return nil, err
	}
	id := c.RuntimeID(uint8(pos[0]&15), uint8(pos[1]), uint8(pos[2]&15), 0)
	c.RUnlock()

	state, ok := block.ByRuntimeID(id)
	if !ok {
		// This should never happen.
		return nil, fmt.Errorf("could not find block state by runtime ID %v", id)
	}
	// TODO: Implement block NBT reading.
	return state, nil
}

// SetBlock writes a block to the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
// An error is returned if the chunk that the block should be written to could not be loaded successfully.
// SetBlock panics if the block passed has not yet been registered using block.Register().
func (w *World) SetBlock(pos block.Position, b block.Block) error {
	runtimeID, ok := block.RuntimeID(b)
	if !ok {
		return fmt.Errorf("runtime ID of block state %+v not found", b)
	}

	c, err := w.chunk(chunkPosFromBlockPos(pos), false)
	if err != nil {
		return err
	}
	c.SetRuntimeID(uint8(pos[0]&15), uint8(pos[1]), uint8(pos[2]&15), 0, runtimeID)
	// TODO: Implement block NBT writing.
	c.Unlock()

	for _, viewer := range w.Viewers(pos.Vec3()) {
		viewer.ViewBlockUpdate(pos, b)
	}
	return nil
}

// BreakBlock breaks a block at the position passed. Unlike when setting the block at that position to air,
// BreakBlock will also show particles.
func (w *World) BreakBlock(pos block.Position) error {
	old, err := w.Block(pos)
	if err != nil {
		return fmt.Errorf("cannot get block at position broken: %v", err)
	}
	_ = w.SetBlock(pos, block.Air{})
	w.AddParticle(pos.Vec3().Add(mgl32.Vec3{0.5, 0.5, 0.5}), particle.BlockBreak{Block: old})
	return nil
}

// Time returns the current time of the world. The time is incremented every 1/20th of a second, unless
// World.StopTime() is called.
func (w *World) Time() int {
	return int(atomic.LoadInt64(&w.time))
}

// SetTime sets the new time of the world. SetTime will always work, regardless of whether the time is stopped
// or not.
func (w *World) SetTime(new int) {
	atomic.StoreInt64(&w.time, int64(new))
	for _, viewer := range w.allViewers() {
		viewer.ViewTime(new)
	}
}

// StopTime stops the time in the world. When called, the time will no longer cycle and the world will remain
// at the time when StopTime is called. The time may be restarted by calling World.StartTime().
// StopTime will not do anything if the time is already stopped.
func (w *World) StopTime() {
	atomic.StoreUint32(&w.timeStopped, 1)
}

// StartTime restarts the time in the world. When called, the time will start cycling again and the day/night
// cycle will continue. The time may be stopped again by calling World.StopTime().
// StartTime will not do anything if the time is already started.
func (w *World) StartTime() {
	atomic.StoreUint32(&w.timeStopped, 0)
}

// AddParticle spawns a particle at a given position in the world. Viewers that are viewing the chunk will be
// shown the particle.
func (w *World) AddParticle(pos mgl32.Vec3, p particle.Particle) {
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewParticle(pos, p)
	}
}

// entityWorlds holds a list of all entities added to a world. It may be used to lookup the world that an
// entity is currently in.
var entityWorlds = map[Entity]*World{}
var worldsMu sync.RWMutex

// AddEntity adds an entity to the world at the position that the entity has. The entity will be visible to
// all viewers of the world that have the chunk of the entity loaded.
// If the chunk that the entity is in is not yet loaded, it will first be loaded.
// If the entity passed to AddEntity is currently in a world, it is first removed from that world.
func (w *World) AddEntity(e Entity) {
	if e.World() != nil {
		e.World().RemoveEntity(e)
	}
	chunkPos := chunkPosFromVec3(e.Position())
	c, err := w.chunk(chunkPos, true)
	if err != nil {
		w.log.Errorf("error loading chunk to add entity: %v", err)
	}
	c.RUnlock()

	worldsMu.Lock()
	entityWorlds[e] = w
	worldsMu.Unlock()

	w.entityMutex.Lock()
	w.entities[chunkPos] = append(w.entities[chunkPos], e)
	w.entityMutex.Unlock()

	w.viewerMutex.RLock()
	for _, viewer := range w.viewers[chunkPos] {
		// We show the entity to all viewers currently in the chunk that the entity is spawned in.
		showEntity(e, viewer)
	}
	w.viewerMutex.RUnlock()
}

// RemoveEntity removes an entity from the world that is currently present in it. Any viewers of the entity
// will no longer be able to see it.
// RemoveEntity operates assuming the position of the entity is the same as where it is currently in the
// world. If it can not find it there, it will loop through all entities and try to find it.
// RemoveEntity assumes the entity is currently loaded and in a loaded chunk. If not, the function will not do
// anything.
func (w *World) RemoveEntity(e Entity) {
	chunkPos := chunkPosFromVec3(e.Position())

	worldsMu.Lock()
	delete(entityWorlds, e)
	worldsMu.Unlock()

	w.entityMutex.Lock()
	if _, ok := w.chunkCache().Get(chunkPos.Hash()); !ok {
		// The chunk wasn't loaded, so we can't remove any entity from the chunk.
		w.entityMutex.Unlock()
		return
	}
	if !w.removeEntity(chunkPos, e) {

		w.log.Debugf("entity %T cannot be found at chunk position %v: looking for other chunks", e, chunkPos)
		for c := range w.entities {
			// Try to remove the entity from every other chunk until we find it: This is a very heavy
			// operation, but it shouldn't typically occur.
			if w.removeEntity(c, e) {
				break
			}
		}
	}
	w.entityMutex.Unlock()
}

// OfEntity attempts to return a world that an entity is currently in. If the entity was not currently added
// to a world, the world returned is nil and the bool returned is false.
func OfEntity(e Entity) (*World, bool) {
	worldsMu.RLock()
	w, ok := entityWorlds[e]
	worldsMu.RUnlock()
	return w, ok
}

// Spawn returns the spawn of the world. Every new player will by default spawn on this position in the world
// when joining.
func (w *World) Spawn() block.Position {
	return w.provider().WorldSpawn()
}

// SetSpawn sets the spawn of the world to a different position. The player will be spawned in the center of
// this position when newly joining.
func (w *World) SetSpawn(pos block.Position) {
	w.provider().SetWorldSpawn(pos)
}

// LoadDefaultGameMode returns the default game mode of the world. When players join, they are given this game
// mode.
// The default game mode may be changed using SaveDefaultGameMode().
func (w *World) DefaultGameMode() gamemode.GameMode {
	w.gamemodeMu.RLock()
	defer w.gamemodeMu.RUnlock()
	return w.defaultGameMode
}

// SaveDefaultGameMode changes the default game mode of the world. When players join, they are then given that
// game mode.
func (w *World) SetDefaultGameMode(mode gamemode.GameMode) {
	w.gamemodeMu.Lock()
	w.defaultGameMode = mode
	w.gamemodeMu.Unlock()
}

// Provider changes the provider of the world to the provider passed. If nil is passed, the NoIOProvider
// will be set, which does not read or write any data.
func (w *World) Provider(p Provider) {
	w.pMutex.Lock()
	defer w.pMutex.Unlock()

	if p == nil {
		p = NoIOProvider{}
	}
	w.prov = p
	w.name = p.WorldName()
	w.gamemodeMu.Lock()
	w.defaultGameMode = p.LoadDefaultGameMode()
	w.gamemodeMu.Unlock()
	atomic.StoreInt64(&w.time, p.LoadTime())
	if timeRunning := p.LoadTimeCycle(); !timeRunning {
		atomic.StoreUint32(&w.timeStopped, 1)
	}
	w.initChunkCache()
}

// Generator changes the generator of the world to the one passed. If nil is passed, the generator is set to
// the default: FlatGenerator.
func (w *World) Generator(g Generator) {
	w.gMutex.Lock()
	defer w.gMutex.Unlock()

	if g == nil {
		g = FlatGenerator{}
	}
	w.gen = g
}

// Start changes the current handler of the world. As a result, events called by the world will call
// handlers of the Handler passed.
// Start sets the world's handler to NopHandler if nil is passed.
func (w *World) Handle(h Handler) {
	w.hMutex.Lock()
	defer w.hMutex.Unlock()

	if h == nil {
		h = NopHandler{}
	}
	w.hand = h
}

// Viewers returns a list of all viewers viewing the position passed. A viewer will be assumed to be watching
// if the position is within one of the chunks that the viewer is watching.
func (w *World) Viewers(pos mgl32.Vec3) []Viewer {
	return w.chunkViewers(chunkPosFromVec3(pos))
}

// Close closes the world and saves all chunks currently loaded.
func (w *World) Close() error {
	w.cancelTick()

	w.viewerMutex.Lock()
	w.viewers = map[ChunkPos][]Viewer{}
	w.viewerMutex.Unlock()

	w.log.Debug("Saving chunks in memory to disk...")
	for key := range w.chunkCache().Items() {
		// We delete all chunks from the cache so that they are saved to the provider.
		w.chunkCache().Delete(key)
	}
	w.log.Debug("Updating level.dat values...")
	w.provider().SaveTime(atomic.LoadInt64(&w.time))
	w.provider().SaveTimeCycle(atomic.LoadUint32(&w.timeStopped) == 0)

	w.gamemodeMu.RLock()
	w.provider().SaveDefaultGameMode(w.defaultGameMode)
	w.gamemodeMu.RUnlock()

	w.log.Debug("Writing level.dat...")
	if err := w.provider().Close(); err != nil {
		w.log.Errorf("error closing world provider: %v", err)
	}
	w.Handle(NopHandler{})
	return nil
}

// startTicking starts ticking the world, updating all entities, blocks and other features such as the time of
// the world, as required.
func (w *World) startTicking() {
	ticker := time.NewTicker(time.Second / 20)
	defer ticker.Stop()

	tick := 0
	for {
		select {
		case <-ticker.C:
			w.tick(tick)
			tick++
		case <-w.stopTick.Done():
			// The world was closed, so we should stop ticking.
			return
		}
	}
}

// tick ticks the world and updates the time, blocks and entities that require updates.
func (w *World) tick(tick int) {
	if atomic.LoadUint32(&w.timeStopped) == 0 {
		// Only if the time is not stopped, we add one to the current time.
		atomic.AddInt64(&w.time, 1)
	}
	if tick%20 == 0 {
		for _, viewer := range w.allViewers() {
			viewer.ViewTime(int(atomic.LoadInt64(&w.time)))
		}
	}
	w.tickEntities()
}

// tickEntities ticks all entities in the world, making sure they are still located in the correct chunks and
// updating where necessary.
func (w *World) tickEntities() {
	w.entityMutex.Lock()
	for chunkPos, entities := range w.entities {
		chunkEntities := make([]Entity, 0, len(entities))
		for _, entity := range entities {
			// The entity was stored using an outdated chunk position. We update it and make sure it is ready
			// for viewers to view it.
			newChunkPos := chunkPosFromVec3(entity.Position())
			if newChunkPos != chunkPos {
				w.entities[newChunkPos] = append(w.entities[newChunkPos], entity)

				w.viewerMutex.RLock()
				for _, viewer := range w.viewers[chunkPos] {
					if !w.hasViewer(newChunkPos, viewer) {
						// First we hide the entity from all viewers that were previously viewing it, but no
						// longer are.
						viewer.HideEntity(entity)
					}
				}
				for _, viewer := range w.viewers[newChunkPos] {
					if !w.hasViewer(chunkPos, viewer) {
						// Then we show the entity to all viewers that are now viewing the entity in the new
						// chunk.
						showEntity(entity, viewer)
					}
				}
				w.viewerMutex.RUnlock()
				continue
			}
			chunkEntities = append(chunkEntities, entity)
		}
		if len(chunkEntities) == 0 {
			// There are no more entities stored in this chunk: We delete the entry from the map.
			delete(w.entities, chunkPos)
			continue
		}
		w.entities[chunkPos] = chunkEntities
	}
	w.entityMutex.Unlock()
}

// removeEntity attempts to remove an entity located in a chunk at the chunk position passed. If found, it
// removes the entity and returns true. If it can't be found, removeEntity returns false.
func (w *World) removeEntity(chunkPos ChunkPos, e Entity) (found bool) {
	n := make([]Entity, 0, len(w.entities[chunkPos]))
	for _, entity := range w.entities[chunkPos] {
		if entity != e {
			n = append(n, entity)
			continue
		}
		w.viewerMutex.RLock()
		for _, viewer := range w.viewers[chunkPos] {
			viewer.HideEntity(e)
		}
		w.viewerMutex.RUnlock()
		found = true
	}
	if len(n) == 0 {
		// The entity is the last in the chunk, so we can delete the value from the map.
		delete(w.entities, chunkPos)
		return
	}
	w.entities[chunkPos] = n
	return
}

// addViewer adds a viewer to the world at a given position. Any events that happen in the chunk at that
// position, such as block changes, entity changes etc., will be sent to the viewer.
func (w *World) addViewer(pos ChunkPos, viewer Viewer) {
	w.viewerMutex.Lock()
	w.viewers[pos] = append(w.viewers[pos], viewer)
	w.viewerMutex.Unlock()

	// After adding the viewer to the chunk, we also need to send all entities currently in the chunk that the
	// viewer is added to.
	w.entityMutex.RLock()
	for _, entity := range w.entities[pos] {
		showEntity(entity, viewer)
	}
	w.entityMutex.RUnlock()

	viewer.ViewTime(w.Time())
}

// removeViewer removes a viewer from the world at a given position. All entities will be hidden from the
// viewer and no more calls will be made when events in the chunk happen.
func (w *World) removeViewer(pos ChunkPos, viewer Viewer) {
	w.viewerMutex.Lock()
	n := make([]Viewer, 0, len(w.viewers[pos]))
	for _, v := range w.viewers[pos] {
		if v != viewer {
			// Add all viewers but the one to remove to the new viewers slice.
			n = append(n, v)
		}
	}
	if len(n) == 0 {
		delete(w.viewers, pos)
	} else {
		w.viewers[pos] = n
	}
	w.viewerMutex.Unlock()

	// After removing the viewer from the chunk, we also need to hide all entities from the viewer.
	w.entityMutex.RLock()
	for _, entity := range w.entities[pos] {
		viewer.HideEntity(entity)
	}
	w.entityMutex.RUnlock()
}

// hasViewer checks if a chunk at a particular chunk position has the viewer passed. If so, true is returned.
func (w *World) hasViewer(pos ChunkPos, viewer Viewer) bool {
	for _, v := range w.viewers[pos] {
		if v == viewer {
			return true
		}
	}
	return false
}

// allViewers returns a list of all viewers of the world, regardless of where in the world they are viewing.
func (w *World) allViewers() []Viewer {
	var v []Viewer
	found := make(map[Viewer]struct{})
	w.viewerMutex.RLock()
	for _, c := range w.viewers {
		for _, viewer := range c {
			if _, ok := found[viewer]; ok {
				// We've already found this viewer in another chunk. Don't add it again.
				continue
			}
			found[viewer] = struct{}{}
			v = append(v, viewer)
		}
	}
	w.viewerMutex.RUnlock()
	return v
}

// provider returns the provider of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) provider() Provider {
	w.pMutex.RLock()
	provider := w.prov
	w.pMutex.RUnlock()
	return provider
}

// handler returns the handler of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) handler() Handler {
	w.hMutex.RLock()
	handler := w.hand
	w.hMutex.RUnlock()
	return handler
}

// generator returns the generator of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) generator() Generator {
	w.gMutex.RLock()
	generator := w.gen
	w.gMutex.RUnlock()
	return generator
}

// chunkCache returns the chunk cache of the world. It should always be used, rather than direct field
// access, in order to provide synchronisation safety.
func (w *World) chunkCache() *cache.Cache {
	w.pMutex.RLock()
	c := w.cCache
	w.pMutex.RUnlock()
	return c
}

// chunkViewers returns a list of all viewers of a chunk at a given position.
func (w *World) chunkViewers(pos ChunkPos) []Viewer {
	w.viewerMutex.RLock()
	viewers := make([]Viewer, len(w.viewers[pos]))
	copy(viewers, w.viewers[pos])
	w.viewerMutex.RUnlock()
	return viewers
}

// moveChunkEntity moves an entity from one chunk to another. It makes sure viewers of the old chunk that are
// not viewing the new one no longer see the entity, and viewers of the new chunk that were not already
// viewing the old chunk are shown the entity.
func (w *World) moveChunkEntity(e Entity, chunkPos, newChunkPos ChunkPos) {
	w.entityMutex.Lock()
	n := make([]Entity, 0, len(w.entities[chunkPos]))
	for _, entity := range w.entities[chunkPos] {
		if entity != e {
			n = append(n, entity)
		}
	}
	if len(n) == 0 {
		// The entity is the last in the chunk, so we can delete the value from the map.
		delete(w.entities, chunkPos)
	} else {
		w.entities[chunkPos] = n
	}

	w.entities[newChunkPos] = append(w.entities[newChunkPos], e)
	w.entityMutex.Unlock()

	w.viewerMutex.RLock()
	for _, viewer := range w.viewers[chunkPos] {
		if !w.hasViewer(newChunkPos, viewer) {
			// First we hide the entity from all viewers that were previously viewing it, but no longer are.
			viewer.HideEntity(e)
		}
	}
	for _, viewer := range w.viewers[newChunkPos] {
		if !w.hasViewer(chunkPos, viewer) {
			// Then we show the entity to all viewers that are now viewing the entity in the new chunk.
			showEntity(e, viewer)
		}
	}
	w.viewerMutex.RUnlock()
}

// showEntity shows an entity to a viewer of the world. It makes sure everything of the entity, including the
// items held, is shown.
func showEntity(e Entity, viewer Viewer) {
	viewer.ViewEntity(e)
	viewer.ViewEntityState(e, e.State())
	viewer.ViewEntityItems(e)
}

// chunk reads a chunk from the position passed. If a chunk at that position is not yet loaded, the chunk is
// loaded from the provider, or generated if it did not yet exist. Both of these actions are done
// synchronously.
// An error is returned if the chunk could not be loaded successfully.
// chunk locks the chunk returned, meaning that any call to chunk made at the same time has to wait until the
// user calls Chunk.Unlock() on the chunk returned.
func (w *World) chunk(pos ChunkPos, readOnly bool) (c *chunk.Chunk, err error) {
	s, ok := w.chunkCache().Get(pos.Hash())
	if !ok {
		// We don't currently have the chunk cached, so we have to load it from the provider.
		var found bool
		c, found, err = w.provider().LoadChunk(pos)
		if err != nil {
			return nil, fmt.Errorf("error loading chunk %v: %v", pos, err)
		}
		if !found {
			// The provider doesn't have a chunk saved at this position, so we generate a new one.
			c = &chunk.Chunk{}
			w.generator().GenerateChunk(pos, c)
		} else {
			entities, err := w.provider().LoadEntities(pos)
			if err != nil {
				return nil, fmt.Errorf("error loading entities of chunk %v: %v", pos, err)
			}
			if len(entities) != 0 {
				w.entityMutex.Lock()
				w.entities[pos] = entities
				w.entityMutex.Unlock()
			}
		}
	} else {
		c = s.(*chunk.Chunk)
	}
	// We set the chunk back to the cache right away, so that the expiration time is reset.
	w.chunkCache().Set(pos.Hash(), c, cache.DefaultExpiration)
	w.chunkCache().Set(pos.timeHash(), time.Now().Add(time.Minute*5), cache.DefaultExpiration)

	if readOnly {
		c.RLock()
	} else {
		c.Lock()
	}
	return c, nil
}

// saveChunk is called when a chunk is removed from the cache. We first compact the chunk, then we write it to
// the provider.
func (w *World) saveChunk(hash string, i interface{}) {
	if _, ok := i.(*chunk.Chunk); !ok {
		return
	}
	pos := chunkPosFromHash(hash)

	c := i.(*chunk.Chunk)
	c.Lock()
	c.Compact()
	c.Unlock()

	if err := w.provider().SaveChunk(pos, c); err != nil {
		w.log.Errorf("error saving chunk %v to provider: %v", pos, err)
	}
	w.entityMutex.Lock()
	entities := w.entities[pos]
	delete(w.entities, pos)
	w.entityMutex.Unlock()

	if err := w.provider().SaveEntities(pos, entities); err != nil {
		w.log.Errorf("error saving entities in chunk %v to provider: %v", pos, err)
	}
	for _, entity := range entities {
		_ = entity.Close()
	}
}

// initChunkCache initialises the chunk cache of the world to its default values.
func (w *World) initChunkCache() {
	w.cCache = cache.New(cache.NoExpiration, cache.NoExpiration)
	w.cCache.OnEvicted(w.saveChunk)
	go func() {
		t := time.NewTicker(time.Minute * 5)
		for {
			select {
			case <-t.C:
				for k, i := range w.cCache.Items() {
					if len(k) == 8 {
						// A chunk was stored at this hash, but we're looking for times.
						continue
					}
					pos := chunkPosFromHash(k)
					if len(w.chunkViewers(pos)) != 0 {
						// There are still viewers viewing the chunk: Don't evict it.
						w.chunkCache().Set(k, time.Now().Add(time.Minute*5), cache.DefaultExpiration)
						continue
					}
					if i.Object.(time.Time).Sub(time.Now()) <= 0 {
						// The time set is below the current time: We should evict the chunk.
						w.cCache.Delete(k)
						w.cCache.Delete(k[:8])
					}
				}
			case <-w.stopTick.Done():
				return
			}
		}
	}()
}
