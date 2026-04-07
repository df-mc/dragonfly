package world

import (
	"github.com/df-mc/dragonfly/server/world/chunk"
)

// Generator handles the generating of newly created chunks. Worlds have one generator which is used to
// generate chunks when the provider of the world cannot find a chunk at a given chunk position.
type Generator interface {
	// GenerateChunk generates a chunk at a chunk position passed. The generator sets blocks in the chunk that
	// is passed to the method.
	GenerateChunk(pos ChunkPos, chunk *chunk.Chunk)
}

// NopGenerator is the default generator a world. It places no blocks in the world which results in a void
// world.
type NopGenerator struct{}

// GenerateChunk ...
func (NopGenerator) GenerateChunk(ChunkPos, *chunk.Chunk) {}

// chunkRequest ...
type chunkRequest struct {
	pos        ChunkPos
	callbacks  []chunkCallback
	generating bool

	close  chan struct{}
	col    *chunk.Column
	result *Column
}

// Do adds callback to list of all callbacks.
func (r *chunkRequest) Do(tx *Tx, receiver chunkCallback) {
	r.callbacks = append(r.callbacks, receiver)
	if !r.generating {
		r.generating = true
		w := tx.World()
		go r.load(w)
	}
}

// doImmediate waits till chunk is loaded and returns it.
func (r *chunkRequest) doImmediate(tx *Tx) *Column {
	<-r.close
	r.signal(tx)
	return r.result
}

// load loads chunk or generates it.
func (r *chunkRequest) load(w *World) {
	r.col = w.loadChunk(r.pos)
	w.Exec(r.signal)
	close(r.close)
}

// signal calls all callbacks and adds chunk to the world.
func (r *chunkRequest) signal(tx *Tx) {
	if r.result != nil {
		return
	}
	w := tx.World()
	pos := r.pos

	delete(w.chunkRequests, pos)
	r.result = w.addChunk(pos, r.col)
	for _, recv := range r.callbacks {
		recv(tx, r.result)
	}
}

type chunkCallback = func(tx *Tx, chunk *Column)
