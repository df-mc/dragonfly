package world

import "github.com/df-mc/dragonfly/server/world/chunk"

// chunkRequest tracks one asynchronous chunk load for a specific position and
// collects all callbacks waiting for that chunk.
type chunkRequest struct {
	pos        ChunkPos
	callbacks  []chunkCallback
	generating bool

	close  chan struct{}
	col    *chunk.Column
	result *Column
}

// defaultChunkLoadWorkers is the number of asynchronous chunk load workers
// started when Config.ChunkLoadWorkers is not set.
const defaultChunkLoadWorkers = 4

// chunkCallback is called on the transaction path after a chunk is installed.
type chunkCallback = func(tx *Tx, chunk *Column)

// chunkRequestHandler dispatches asynchronous chunk requests to a loading strategy.
type chunkRequestHandler interface {
	handleChunkRequest(*chunkRequest)
}

// asyncChunkRequestHandler dispatches chunk requests to a bounded worker queue.
type asyncChunkRequestHandler struct {
	w     *World
	queue chan *chunkRequest
}

func newAsyncChunkRequestHandler(w *World) *asyncChunkRequestHandler {
	return &asyncChunkRequestHandler{w: w, queue: make(chan *chunkRequest, 4096)}
}

// Do registers receiver to be called when the chunk is loaded. The first call
// starts the asynchronous load.
func (r *chunkRequest) Do(tx *Tx, receiver chunkCallback) {
	r.callbacks = append(r.callbacks, receiver)
	if !r.generating {
		r.generating = true
		tx.World().chunkRequestHandler.handleChunkRequest(r)
	}
}

// doImmediate waits until the chunk is loaded and returns it.
func (r *chunkRequest) doImmediate(tx *Tx) *Column {
	<-r.close
	r.signal(tx)
	return r.result
}

// load reads or generates the chunk, then schedules installation back onto the
// world transaction queue.
func (r *chunkRequest) load(w *World) {
	r.col = w.loadChunk(r.pos)
	close(r.close)
	select {
	case <-w.closing:
		return
	default:
		w.Exec(r.signal)
	}
}

// handleChunkRequest queues r on the bounded asynchronous chunk loading pool.
func (h *asyncChunkRequestHandler) handleChunkRequest(r *chunkRequest) {
	select {
	case h.queue <- r:
	case <-h.w.closing:
		close(r.close)
	}
}

// handle processes the chunk load queue until the world closes.
func (h *asyncChunkRequestHandler) handle() {
	defer h.w.running.Done()
	for {
		select {
		case <-h.w.closing:
			return
		default:
		}
		select {
		case r := <-h.queue:
			r.load(h.w)
		case <-h.w.closing:
			return
		}
	}
}

// signal installs the loaded chunk and invokes all waiting callbacks.
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
