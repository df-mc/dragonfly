package world

import (
	"sync"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/world/chunk"
)

// chunkRequest tracks a chunk that is being loaded or generated in the
// background. All callers waiting for the same chunk share a single request.
type chunkRequest struct {
	pos       ChunkPos
	callbacks []chunkCallback
	signalled bool

	mu     sync.Mutex
	state  chunkRequestState
	queued *atomic.Int32

	done   chan struct{}
	col    *chunk.Column
	err    error
	result *Column
}

// chunkRequestState distinguishes unclaimed work from loads and skipped requests.
type chunkRequestState uint8

const (
	chunkRequestQueued chunkRequestState = iota
	chunkRequestLoading
	chunkRequestSkipped
)

// defaultChunkLoadWorkers is the number of chunk load workers started when
// Config.ChunkLoadWorkers is not set.
const defaultChunkLoadWorkers = 1

// chunkCallback tracks a caller's interest and delivers its chunk on the world goroutine.
type chunkCallback struct {
	wanted  *atomic.Bool
	receive func(tx *Tx, col *Column)
}

// chunkWorkerPool runs chunk requests on a fixed number of background workers.
type chunkWorkerPool struct {
	w     *World
	queue chan *chunkRequest
	wg    sync.WaitGroup

	mu     sync.Mutex
	closed bool
}

// newChunkWorkerPool creates the workers' bounded waiting queue.
func newChunkWorkerPool(w *World) *chunkWorkerPool {
	return &chunkWorkerPool{w: w, queue: make(chan *chunkRequest, 4096)}
}

// doImmediate claims queued work or waits for its active load, then installs it.
func (r *chunkRequest) doImmediate(tx *Tx) *Column {
	if r.take(chunkRequestLoading) {
		r.col, r.err = tx.World().loadChunk(r.pos)
		close(r.done)
	}
	<-r.done
	r.signal(tx)
	return r.result
}

// load loads or generates the chunk and hands it back to the world to be
// added.
func (r *chunkRequest) load(w *World) {
	owned, needed := r.begin()
	if !owned {
		return
	}
	if needed {
		r.col, r.err = w.loadChunk(r.pos)
	}
	close(r.done)
	w.Do(r.signal)
}

// leave releases the creator's waiting allowance. r.mu must be held.
func (r *chunkRequest) leave(state chunkRequestState) {
	r.state = state
	if r.queued != nil {
		r.queued.Add(-1)
	}
}

// take claims an unstarted request for an immediate load or cancellation.
func (r *chunkRequest) take(state chunkRequestState) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != chunkRequestQueued {
		return false
	}
	r.leave(state)
	return true
}

// begin claims a request for a worker, skipping it if every caller withdrew.
func (r *chunkRequest) begin() (owned, needed bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state != chunkRequestQueued {
		return false, false
	}
	for _, callback := range r.callbacks {
		if callback.wanted == nil || callback.wanted.Load() {
			r.leave(chunkRequestLoading)
			return true, true
		}
	}
	r.leave(chunkRequestSkipped)
	return true, false
}

// attach registers a caller unless this request has already been skipped.
func (r *chunkRequest) attach(callback chunkCallback) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state == chunkRequestSkipped {
		return false
	}
	r.callbacks = append(r.callbacks, callback)
	return true
}

// abort releases a queued request when the world closes without stopping active loads.
func (r *chunkRequest) abort() {
	if r.take(chunkRequestSkipped) {
		close(r.done)
	}
}

// schedule hands r to the workers without blocking. It returns false if the
// queue is full, the loader has two batches waiting, or the world is closing.
func (p *chunkWorkerPool) schedule(r *chunkRequest, batchSize int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.w.closed.Load() {
		p.closed = true
		return false
	}
	// Leave room for a full batch while the previous batch is still waiting.
	// The physical capacity bounds the multiplication for arbitrarily large batches.
	limit := 2 * min(batchSize, cap(p.queue))
	if r.queued != nil {
		if int(r.queued.Load()) >= limit {
			return false
		}
		r.queued.Add(1)
	}
	select {
	case p.queue <- r:
		return true
	default:
		if r.queued != nil {
			r.queued.Add(-1)
		}
		return false
	}
}

// handle continuously processes chunk requests until the world starts closing.
func (p *chunkWorkerPool) handle() {
	defer p.wg.Done()
	for {
		if p.w.closed.Load() {
			p.drainAndAbort()
			return
		}
		select {
		case r := <-p.queue:
			r.load(p.w)
		case <-p.w.closeStarted:
			p.drainAndAbort()
			return
		}
	}
}

// drainAndAbort cancels all remaining requests and stops accepting new ones.
func (p *chunkWorkerPool) drainAndAbort() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	for {
		select {
		case r := <-p.queue:
			r.abort()
		default:
			return
		}
	}
}

// signal adds the finished chunk to the world and calls all callers waiting
// for it. It always runs inside a world transaction.
func (r *chunkRequest) signal(tx *Tx) {
	if r.signalled {
		return
	}
	r.signalled = true

	w := tx.World()
	pos := r.pos

	if w.chunkRequests[pos] == r {
		delete(w.chunkRequests, pos)
	}
	if w.closed.Load() {
		return
	}
	if r.err != nil || r.col == nil {
		if r.err != nil {
			w.conf.Log.Error("load chunk: "+r.err.Error(), "X", pos[0], "Z", pos[1])
		}
		for _, recv := range r.callbacks {
			recv.receive(tx, nil)
		}
		return
	}
	r.result = w.addChunk(pos, r.col)
	if w.closed.Load() {
		return
	}
	for _, recv := range r.callbacks {
		recv.receive(tx, r.result)
	}
}
