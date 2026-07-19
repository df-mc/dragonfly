package world

import (
	"sync"

	"github.com/df-mc/dragonfly/server/world/chunk"
)

// chunkRequest tracks a chunk that is being loaded or generated in the
// background. Everyone waiting for the same chunk shares a single request.
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

// chunkCallback is called with a chunk once it has been added to the world.
type chunkCallback = func(tx *Tx, chunk *Column)

// chunkRequestHandler decides how chunk requests are carried out.
type chunkRequestHandler interface {
	handleChunkRequest(*chunkRequest) bool
}

// workerPoolChunkRequestHandler runs chunk requests on a fixed number of
// background workers.
type workerPoolChunkRequestHandler struct {
	w      *World
	queue  chan *chunkRequest
	mu     sync.Mutex
	closed bool
}

func newWorkerPoolChunkRequestHandler(w *World) *workerPoolChunkRequestHandler {
	return &workerPoolChunkRequestHandler{w: w, queue: make(chan *chunkRequest, 4096)}
}

// Do calls receiver once the chunk is ready. The first call starts the
// request; later calls simply wait for the same chunk. Do returns false if the
// request could not be queued, e.g. because the world is closing.
func (r *chunkRequest) Do(tx *Tx, receiver chunkCallback) bool {
	r.callbacks = append(r.callbacks, receiver)
	if !r.queued {
		r.queued = true
		return tx.World().chunkRequestHandler.handleChunkRequest(r)
	}
	return true
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
	go r.queueSignal(w)
}

func (r *chunkRequest) queueSignal(w *World) {
	select {
	case w.queue <- normalTransaction{c: make(chan struct{}), f: r.signal}:
	case <-w.closing:
		return
	}
}

// abort cancels a request that will never be carried out because the world is
// closing, releasing anyone waiting on it.
func (r *chunkRequest) abort() {
	r.aborted = true
	close(r.done)
}

// handleChunkRequest hands r to the workers. It returns false, without
// blocking, if the workers cannot accept the request, e.g. because their queue
// is full or the world is closing. The caller may simply retry later.
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

// handle continuously processes chunk requests until the world starts closing.
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

// drainAndAbort cancels all remaining requests and stops accepting new ones.
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

// signal adds the finished chunk to the world and calls everyone waiting for
// it. It always runs inside a world transaction.
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
