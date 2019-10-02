package world

import (
	"errors"
	"fmt"
	"github.com/dragonfly-tech/dragonfly/dragonfly/block/encoder"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/chunk"
	"github.com/patrickmn/go-cache"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
type World struct {
	name string
	log  *logrus.Logger

	hMutex sync.RWMutex
	h      Handler

	pMutex sync.RWMutex
	p      Provider
	cCache *cache.Cache

	gMutex sync.RWMutex
	g      Generator
}

// New creates a new initialised world. The world may be used right away, but it will not be saved or loaded
// from files until it has been given a different provider than the default. (NoIOProvider)
// By default, the name of the world will be 'World'.
func New(log *logrus.Logger) *World {
	w := &World{
		name:   "World",
		p:      NoIOProvider{},
		g:      FlatGenerator{},
		cCache: cache.New(3*time.Minute, 5*time.Minute),
		log:    log,
	}
	w.cCache.OnEvicted(func(s string, i interface{}) {
		// This function is called when a chunk is removed from the cache. We first compact the chunk, then we
		// write it to the provider.
		c := i.(*chunk.Chunk)
		c.Compact()
		if err := w.p.SaveChunk(ChunkPosFromHash(s), c); err != nil {
			log.Errorf("error saving chunk to provider: %v", err)
		}
	})
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
	id := c.RuntimeID(uint8(pos[0]&15), uint8(pos[1]&15), uint8(pos[2]&15), 0)
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
	c.SetRuntimeID(uint8(pos[0]&15), uint8(pos[1]&15), uint8(pos[2]&15), 0, encoder.RuntimeIDs[protocol.BlockEntry{
		Name: id,
		Data: meta,
	}])
	// TODO: Implement block NBT writing.
	_ = nbt
	c.Unlock()

	return nil
}

// Provider changes the provider of the world to the provider passed. If nil is passed, the NoIOProvider
// will be set, which does not read or write any data.
func (w *World) Provider(p Provider) {
	w.pMutex.Lock()
	defer w.pMutex.Unlock()

	if p == nil {
		p = NoIOProvider{}
	}
	w.p = p
	w.name = p.WorldName()
	w.cCache = cache.New(time.Second*3, time.Second*6)
}

// Generator changes the generator of the world to the one passed. If nil is passed, the generator is set to
// the default: FlatGenerator.
func (w *World) Generator(g Generator) {
	w.gMutex.Lock()
	defer w.gMutex.Unlock()

	if g == nil {
		g = FlatGenerator{}
	}
	w.g = g
}

// Handle changes the current handler of the world. As a result, events called by the world will call
// handlers of the Handler passed.
// Handle sets the world's handler to NopHandler if nil is passed.
func (w *World) Handle(h Handler) {
	w.hMutex.Lock()
	defer w.hMutex.Unlock()

	if h == nil {
		h = NopHandler{}
	}
	w.h = h
}

// Close closes the world and saves all chunks currently loaded.
func (w *World) Close() error {
	for key := range w.cCache.Items() {
		// We delete all chunks from the cache so that they are saved to the provider.
		w.cCache.Delete(key)
	}
	return nil
}

// provider returns the provider of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) provider() Provider {
	w.pMutex.RLock()
	provider := w.p
	w.pMutex.RUnlock()
	return provider
}

// handler returns the handler of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) handler() Handler {
	w.hMutex.RLock()
	handler := w.h
	w.hMutex.RUnlock()
	return handler
}

// generator returns the generator of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) generator() Generator {
	w.gMutex.RLock()
	generator := w.g
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
			return nil, fmt.Errorf("error loading chunk: %v", err)
		}
		if !found {
			// The provider doesn't have a chunk saved at this position, so we generate a new one.
			c = &chunk.Chunk{}
			w.generator().GenerateChunk(pos, c)
		}
	} else {
		c = s.(*chunk.Chunk)
	}
	// We set the chunk back to the cache right away, so that the expiration time is reset.
	w.chunkCache().Set(pos.Hash(), c, cache.DefaultExpiration)

	c.Lock()
	return c, nil
}
