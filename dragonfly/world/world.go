package world

import (
	"errors"
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/block/encoder"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/patrickmn/go-cache"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
// World generally provides a synchronised state: All entities, blocks and players usually operate in this
// world, so World ensures that all its methods will always be safe for simultaneous calls.
type World struct {
	name string
	log  *logrus.Logger

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
	w := &World{
		name:     "World",
		prov:     NoIOProvider{},
		gen:      FlatGenerator{},
		log:      log,
		viewers:  make(map[ChunkPos][]Viewer),
		entities: make(map[ChunkPos][]Entity),
	}
	w.initChunkCache()
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
func (w *World) Block(pos BlockPos) (Block, error) {
	c, err := w.chunk(pos.ChunkPos())
	if err != nil {
		return nil, err
	}
	id := c.RuntimeID(uint8(pos[0]&15), uint8(pos[1]), uint8(pos[2]&15), 0)
	c.Unlock()

	state := encoder.Blocks[id]
	e, ok := encoder.ByID(state.Name)
	if !ok {
		return nil, errors.New("no decoder for " + state.Name)
	}
	// TODO: Implement block NBT reading.
	return e.DecodeBlock(state.Name, state.Data, nil).(Block), nil
}

// SetBlock writes a block to the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
// An error is returned if the chunk that the block should be written to could not be loaded successfully.
// SetBlock panics if the block passed does not have an encoder registered for it in the encoder package.
func (w *World) SetBlock(pos BlockPos, block Block) error {
	e, ok := encoder.ByBlock(block)
	if !ok {
		panic("no encoder for block " + block.Name())
	}
	id, meta, nbt := e.EncodeBlock(block)

	c, err := w.chunk(pos.ChunkPos())
	if err != nil {
		return err
	}
	c.SetRuntimeID(uint8(pos[0]&15), uint8(pos[1]), uint8(pos[2]&15), 0, encoder.RuntimeIDs[protocol.BlockEntry{
		Name: id,
		Data: meta,
	}])
	// TODO: Implement block NBT writing.
	_ = nbt
	c.Unlock()

	return nil
}

// AddEntity adds an entity to the world at the position that the entity has. The entity will be visible to
// all viewers of the world that have the chunk of the entity loaded.
// If the chunk that the entity is in is not yet loaded, it will first be loaded.
func (w *World) AddEntity(e Entity) {
	chunkPos := ChunkPosFromVec3(e.Position())
	c, err := w.chunk(chunkPos)
	if err != nil {
		w.log.Errorf("error loading chunk to add entity: %v", err)
	}
	c.Unlock()

	w.entityMutex.Lock()
	w.entities[chunkPos] = append(w.entities[chunkPos], e)
	w.entityMutex.Unlock()

	w.viewerMutex.RLock()
	for _, viewer := range w.viewers[chunkPos] {
		// We show the entity to all viewers currently in the chunk that the entity is spawned in.
		viewer.ViewEntity(e)
	}
	w.viewerMutex.RUnlock()
}

// RemoveEntity removes an entity from the world that is currently present in it. Any viewers of the entity
// will no longer be able to see it.
// RemoveEntity operates assuming the position of the entity is the same as where it is currently in the
// world. If it can not find it there, it will loop through all entities and try to find it.
func (w *World) RemoveEntity(e Entity) {
	chunkPos := ChunkPosFromVec3(e.Position())

	w.entityMutex.Lock()
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

// MoveEntity moves an entity from one position to another in the world, by adding the delta passed to the
// current position of the entity. It is equivalent to calling entity.Move().
func (w *World) MoveEntity(e Entity, delta mgl32.Vec3) {
	chunkPos := ChunkPosFromVec3(e.Position())
	newChunkPos := ChunkPosFromVec3(e.Position().Add(delta))

	if chunkPos != newChunkPos {
		// The entity moved from one chunk into another, so we need to move it and show it to the new viewers.
		// Old viewers also need to stop viewing this entity.
		w.moveChunkEntity(e, chunkPos, newChunkPos)
	}

	w.viewerMutex.RLock()
	for _, viewer := range w.viewers[newChunkPos] {
		// Finally we show the movement to all viewers of the entity.
		viewer.ViewEntityMovement(e, delta, 0, 0)
	}
	w.viewerMutex.RUnlock()

	// Make sure to set the final position of the entity: It should not yet be applied when making the viewers
	// view the movement.
	e.setPosition(e.Position().Add(delta))
}

// RotateEntity rotates an entity in the position, adding deltaYaw and deltaPitch to the respective values. It
// is equivalent to calling entity.Rotate().
func (w *World) RotateEntity(e Entity, deltaYaw, deltaPitch float32) {
	chunkPos := ChunkPosFromVec3(e.Position())

	w.viewerMutex.RLock()
	for _, viewer := range w.viewers[chunkPos] {
		viewer.ViewEntityMovement(e, mgl32.Vec3{}, deltaYaw, deltaPitch)
	}
	w.viewerMutex.RUnlock()

	e.setYaw(e.Yaw() + deltaYaw)
	e.setPitch(e.Pitch() + deltaPitch)
}

// Entities returns a list of all entities in the world. Note that this includes only entities of loaded
// chunks: Entities in chunks that have not been loaded will not be returned.
func (w *World) Entities() []Entity {
	w.entityMutex.RLock()
	// Make an estimate of about 10 entities per loaded chunk.
	m := make([]Entity, 0, len(w.entities)*10)
	for _, e := range w.entities {
		m = append(m, e...)
	}
	w.entityMutex.RUnlock()
	return m
}

// Spawn returns the spawn of the world. Every new player will by default spawn on this position in the world
// when joining.
func (w *World) Spawn() mgl32.Vec3 {
	return w.provider().WorldSpawn().Vec3()
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

// Close closes the world and saves all chunks currently loaded.
func (w *World) Close() error {
	w.viewerMutex.Lock()
	w.viewers = map[ChunkPos][]Viewer{}
	w.viewerMutex.Unlock()

	for key := range w.cCache.Items() {
		// We delete all chunks from the cache so that they are saved to the provider.
		w.cCache.Delete(key)
	}
	if err := w.provider().Close(); err != nil {
		w.log.Errorf("error closing world provider: %v", err)
	}
	w.Handle(NopHandler{})
	return nil
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
		viewer.ViewEntity(entity)
	}
	w.entityMutex.RUnlock()
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
			viewer.ViewEntity(e)
		}
	}
	w.viewerMutex.RUnlock()
}

// chunk reads a chunk from the position passed. If a chunk at that position is not yet loaded, the chunk is
// loaded from the provider, or generated if it did not yet exist. Both of these actions are done
// synchronously.
// An error is returned if the chunk could not be loaded successfully.
// chunk locks the chunk returned, meaning that any call to chunk made at the same time has to wait until the
// user calls Chunk.Unlock() on the chunk returned.
func (w *World) chunk(pos ChunkPos) (c *chunk.Chunk, err error) {
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

	c.Lock()
	return c, nil
}

// saveChunk is called when a chunk is removed from the cache. We first compact the chunk, then we write it to
// the provider.
func (w *World) saveChunk(hash string, i interface{}) {
	pos := ChunkPosFromHash(hash)

	w.viewerMutex.RLock()
	if len(w.viewers[pos]) != 0 {
		// There are still viewers watching the chunk, so we don't save it and put it back.
		w.chunkCache().Set(hash, i, cache.DefaultExpiration)
		w.viewerMutex.RUnlock()
		return
	}
	w.viewerMutex.RUnlock()

	c := i.(*chunk.Chunk)
	c.Lock()
	c.Compact()
	c.Unlock()

	if err := w.provider().SaveChunk(pos, c); err != nil {
		w.log.Errorf("error saving chunk %v to provider: %v", pos, err)
	}
	w.entityMutex.Lock()
	for _, entity := range w.entities[pos] {
		_ = entity.Close()
	}
	if err := w.provider().SaveEntities(pos, w.entities[pos]); err != nil {
		w.log.Errorf("error saving entities in chunk %v to provider: %v", pos, err)
	}
	delete(w.entities, pos)
	w.entityMutex.Unlock()
}

// initChunkCache initialises the chunk cache of the world to its default values.
func (w *World) initChunkCache() {
	w.cCache = cache.New(3*time.Minute, 5*time.Minute)
	w.cCache.OnEvicted(w.saveChunk)
}
