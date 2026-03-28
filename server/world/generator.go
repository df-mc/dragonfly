package world

import (
	"sync/atomic"

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

// generationRequest ...
type generationRequest struct {
	pos        ChunkPos
	callbacks  []chunkCallback
	generating bool

	immediateTx atomic.Pointer[Tx]
	col         *Column
}

// Do adds callback to list of all callbacks.
func (r *generationRequest) Do(tx *Tx, receiver chunkCallback) {
	r.callbacks = append(r.callbacks, receiver)
	if !r.generating {
		r.generating = true
		w := tx.World()
		go r.generate(w)
	}
}

// doImmediate adds callback and waits till chunk is generated.
func (r *generationRequest) doImmediate(tx *Tx) *Column {
	wait := make(chan struct{})
	r.immediateTx.Store(tx)
	r.callbacks = append(r.callbacks, func(*Tx, *Column) {
		close(wait)
	})
	<-wait
	return r.col
}

// generate starts chunk generation.
func (r *generationRequest) generate(w *World) {
	r.col = newColumn(chunk.New(airRID, w.Range()))
	w.conf.Generator.GenerateChunk(r.pos, r.col.Chunk)
	if tx := r.immediateTx.Load(); tx != nil {
		r.signal(tx)
		return
	}
	w.Exec(r.signal)
}

// signal calls all callbacks and adds chunk to the world.
func (r *generationRequest) signal(tx *Tx) {
	w := tx.World()
	pos := r.pos

	// chunks has been generated.
	delete(w.chunkRequests, pos)
	w.addChunk(pos, r.col)
	for _, recv := range r.callbacks {
		recv(tx, r.col)
	}
}

type chunkCallback = func(tx *Tx, chunk *Column)
