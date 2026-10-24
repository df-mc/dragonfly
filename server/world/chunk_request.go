package world

import (
	"sync"

	"github.com/df-mc/dragonfly/server/world/chunk"
)

// chunkRequest tracks a chunk that is being loaded or generated in the
// background. All callers waiting for the same chunk share a single request.
type chunkRequest struct {
	pos       ChunkPos
	callbacks []chunkCallback
	// loader owns the waiting allowance until a worker starts this request.
	loader    *Loader
	signalled bool

	done   chan struct{}
	col    *chunk.Column
	err    error
	result *Column
}

// defaultChunkLoadWorkers is the number of chunk load workers started when
// Config.ChunkLoadWorkers is not set.
const defaultChunkLoadWorkers = 1

// chunkCallback is called with a chunk once it has been added to the world.
type chunkCallback = func(tx *Tx, col *Column)

// chunkWorkerPool runs chunk requests on a fixed number of background workers.
type chunkWorkerPool struct {
	w     *World
	queue chan *chunkRequest
	wg    sync.WaitGroup

	mu      sync.Mutex
	closed  bool
	waiting map[*Loader]int
}

// newChunkWorkerPool creates the workers' waiting queue. schedule limits
// admission so loaders can reprioritise further requests when they move.
func newChunkWorkerPool(w *World) *chunkWorkerPool {
	return &chunkWorkerPool{w: w, queue: make(chan *chunkRequest, 4096), waiting: make(map[*Loader]int)}
}

// doImmediate blocks until the chunk is ready and returns it.
func (r *chunkRequest) doImmediate(tx *Tx) *Column {
	<-r.done
	r.signal(tx)
	return r.result
}

// load loads or generates the chunk and hands it back to the world to be
// added.
func (r *chunkRequest) load(w *World) {
	r.col, r.err = w.loadChunk(r.pos)
	close(r.done)
	w.Do(r.signal)
}

// abort cancels a request that will never be carried out because the world is
// closing, releasing any callers waiting on it.
func (r *chunkRequest) abort() {
	close(r.done)
}

// schedule hands r to the workers without blocking. It returns false if the
// request cannot be accepted, e.g. when the world is closing. Each loader may
// queue two requests per worker or one caller batch, whichever is larger.
func (p *chunkWorkerPool) schedule(r *chunkRequest, batchSize int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || p.w.closed.Load() {
		p.closed = true
		return false
	}
	if p.waiting[r.loader] >= max(2*p.w.conf.ChunkLoadWorkers, batchSize) {
		return false
	}
	select {
	case p.queue <- r:
		p.waiting[r.loader]++
		return true
	default:
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
			p.mu.Lock()
			p.removeWaiting(r)
			p.mu.Unlock()
			r.load(p.w)
		case <-p.w.closeStarted:
			p.drainAndAbort()
			return
		}
	}
}

// removeWaiting releases the requesting loader's allowance. p.mu must be held.
func (p *chunkWorkerPool) removeWaiting(r *chunkRequest) {
	if p.waiting[r.loader] == 1 {
		delete(p.waiting, r.loader)
	} else {
		p.waiting[r.loader]--
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
			p.removeWaiting(r)
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

	delete(w.chunkRequests, pos)
	if w.closed.Load() {
		return
	}
	if r.err != nil {
		w.conf.Log.Error("load chunk: "+r.err.Error(), "X", pos[0], "Z", pos[1])
		for _, recv := range r.callbacks {
			recv(tx, nil)
		}
		return
	}
	r.result = w.addChunk(pos, r.col)
	if w.closed.Load() {
		return
	}
	for _, recv := range r.callbacks {
		recv(tx, r.result)
	}
}
