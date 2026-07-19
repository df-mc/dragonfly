package world

import (
	"sync"

	"github.com/df-mc/dragonfly/server/world/chunk"
)

// chunkRequest tracks one asynchronous chunk acquisition and collects all
// callbacks waiting for that chunk. Acquisition may load an existing column from
// the provider or generate a new one if the provider returns ErrNotFound.
type chunkRequest struct {
	pos       ChunkPos
	callbacks []chunkCallback
	queued    bool
	aborted   bool
	signalled bool

	done   chan struct{}
	col    *chunk.Column
	err    error
	result *Column
}

// defaultChunkLoadWorkers is the number of asynchronous chunk load workers
// started when Config.ChunkLoadWorkers is not set.
const defaultChunkLoadWorkers = 1

// chunkCallback is called on the transaction path after a chunk is installed.
type chunkCallback = func(tx *Tx, chunk *Column)

// chunkRequestHandler schedules asynchronous chunk acquisition requests.
type chunkRequestHandler interface {
	handleChunkRequest(*chunkRequest) bool
}

// workerPoolChunkRequestHandler processes chunk requests using a bounded worker
// pool.
type workerPoolChunkRequestHandler struct {
	w      *World
	queue  chan *chunkRequest
	mu     sync.Mutex
	closed bool
}

func newWorkerPoolChunkRequestHandler(w *World) *workerPoolChunkRequestHandler {
	return &workerPoolChunkRequestHandler{w: w, queue: make(chan *chunkRequest, 4096)}
}

// Do registers receiver to be called when the chunk is loaded or generated. The
// first call queues the request; later calls only add callbacks. Do returns
// false if the world is closing and the request could not be queued.
func (r *chunkRequest) Do(tx *Tx, receiver chunkCallback) bool {
	r.callbacks = append(r.callbacks, receiver)
	if !r.queued {
		r.queued = true
		return tx.World().chunkRequestHandler.handleChunkRequest(r)
	}
	return true
}

// doImmediate waits until the chunk is loaded or generated and returns it.
func (r *chunkRequest) doImmediate(tx *Tx) *Column {
	<-r.done
	r.signal(tx)
	return r.result
}

// load reads the chunk from the provider or generates a new one, then schedules
// installation back onto the world transaction queue.
func (r *chunkRequest) load(w *World) {
	r.col, r.err = w.loadChunk(r.pos)
	close(r.done)
	go r.queueSignal(w)
}

func (r *chunkRequest) queueSignal(w *World) {
	select {
	case w.queue <- normalTransaction{c: make(chan struct{}), f: r.signal}:
	case <-w.closing:
		return
	}
}

// abort marks r as terminal without a chunk. It is only used during world
// shutdown for queued requests that will never be processed.
func (r *chunkRequest) abort() {
	r.aborted = true
	close(r.done)
}

// handleChunkRequest queues r on the bounded worker pool without blocking the
// world transaction goroutine. The mutex prevents a request from being accepted
// after the pool starts draining during shutdown.
func (h *workerPoolChunkRequestHandler) handleChunkRequest(r *chunkRequest) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return false
	}
	select {
	case <-h.w.closeStarted:
		h.closed = true
		return false
	default:
	}
	select {
	case h.queue <- r:
		return true
	default:
		return false
	}
}

// handle processes the chunk load queue until the world starts closing.
func (h *workerPoolChunkRequestHandler) handle() {
	defer h.w.running.Done()
	for {
		select {
		case <-h.w.closeStarted:
			h.drainAndAbort()
			return
		default:
		}
		select {
		case r := <-h.queue:
			r.load(h.w)
		case <-h.w.closeStarted:
			h.drainAndAbort()
			return
		}
	}
}

// drainAndAbort cancels all queued requests and marks the handler as closed.
func (h *workerPoolChunkRequestHandler) drainAndAbort() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for {
		select {
		case r := <-h.queue:
			r.abort()
		default:
			return
		}
	}
}

// signal installs the loaded or generated chunk and invokes all waiting
// callbacks. It runs on the world transaction queue.
func (r *chunkRequest) signal(tx *Tx) {
	if r.signalled {
		return
	}
	r.signalled = true

	w := tx.World()
	pos := r.pos

	delete(w.chunkRequests, pos)
	if r.aborted {
		return
	}
	select {
	case <-w.closeStarted:
		return
	default:
	}
	if r.err != nil {
		w.conf.Log.Error("load chunk: "+r.err.Error(), "X", pos[0], "Z", pos[1])
		if r.col == nil {
			for _, recv := range r.callbacks {
				recv(tx, nil)
			}
			return
		}
	}
	r.result = w.addChunk(pos, r.col)
	select {
	case <-w.closeStarted:
		return
	default:
	}
	for _, recv := range r.callbacks {
		recv(tx, r.result)
	}
}
