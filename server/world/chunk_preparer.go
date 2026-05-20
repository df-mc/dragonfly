package world

import (
	"errors"
	"sync"

	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
)

type chunkPrepareState struct {
	// token identifies the currently accepted prepare request for a chunk. It lets
	// the transaction goroutine drop stale completions from older requests.
	token uint64
}

type chunkPrepareRequest struct {
	// pos is the chunk position to load from the provider or generate.
	pos ChunkPos
	// token mirrors the pending state token that must match before commit.
	token uint64
}

type chunkPrepareResult struct {
	// pos is the chunk position this result belongs to.
	pos ChunkPos
	// token is checked against pendingChunks before the result is published.
	token uint64
	// column is private prepared chunk data. It is nil when err is non-nil.
	column *chunk.Column
	// err is the provider/generator preparation error, if preparation failed.
	err error
}

type chunkPreparer struct {
	// w is used by workers to access immutable config and to schedule commit transactions.
	w *World

	// requests queues accepted chunk preparation jobs for worker goroutines.
	requests chan chunkPrepareRequest
	// closing is closed to ask workers to stop and to make new requests fail.
	closing chan struct{}
	// closed is closed after all workers have exited.
	closed chan struct{}

	// wg tracks active worker goroutines.
	wg sync.WaitGroup
}

// newChunkPreparer starts a bounded worker pool used to load or generate chunks
// outside the world transaction goroutine. Workers only prepare private chunk
// data; publishing remains the responsibility of the world transaction loop.
func newChunkPreparer(w *World, workers, queueSize int) *chunkPreparer {
	p := &chunkPreparer{
		w:        w,
		requests: make(chan chunkPrepareRequest, queueSize),
		closing:  make(chan struct{}),
		closed:   make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	go func() {
		p.wg.Wait()
		close(p.closed)
	}()
	return p
}

// request enqueues a chunk preparation job without blocking. False means the
// worker queue is full or shutting down, so the caller should not mark the
// chunk as pending.
func (p *chunkPreparer) request(req chunkPrepareRequest) bool {
	select {
	case p.requests <- req:
		return true
	case <-p.closing:
		return false
	default:
		return false
	}
}

// close stops all prepare workers and waits until they have exited. Any
// completions produced after shutdown starts are ignored.
func (p *chunkPreparer) close() {
	select {
	case <-p.closing:
		return
	default:
		close(p.closing)
	}
	<-p.closed
}

func (p *chunkPreparer) worker() {
	defer p.wg.Done()
	for {
		select {
		case req := <-p.requests:
			p.submit(p.prepare(req))
		default:
			select {
			case req := <-p.requests:
				p.submit(p.prepare(req))
			case <-p.closing:
				return
			}
		}
	}
}

// prepare performs the slow provider/generator work using data that is not yet
// visible through World.chunks. The result must be committed on the transaction
// goroutine before it can be used by gameplay or viewers.
func (p *chunkPreparer) prepare(req chunkPrepareRequest) chunkPrepareResult {
	column, err := p.w.conf.Provider.LoadColumn(req.pos, p.w.conf.Dim)
	if err == nil {
		return chunkPrepareResult{pos: req.pos, token: req.token, column: column}
	}
	if !errors.Is(err, leveldb.ErrNotFound) {
		return chunkPrepareResult{pos: req.pos, token: req.token, err: err}
	}

	c := chunk.New(p.w.conf.Blocks, p.w.Range())
	p.w.conf.Generator.GenerateChunk(req.pos, c)
	return chunkPrepareResult{
		pos:   req.pos,
		token: req.token,
		column: &chunk.Column{
			Chunk: c,
		},
	}
}

// submit schedules a prepared chunk for transaction-thread publication. It
// does not wait for the commit, which keeps workers available for more prepare
// work.
func (p *chunkPreparer) submit(result chunkPrepareResult) {
	select {
	case <-p.closing:
		return
	default:
	}
	p.w.Exec(func(tx *Tx) {
		tx.World().commitPreparedChunk(result)
	})
}
