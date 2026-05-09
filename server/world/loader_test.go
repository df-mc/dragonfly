package world

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
)

func TestLoaderLoadAlreadyLoadedChunkDoesNotDeadlock(t *testing.T) {
	w := Config{SaveInterval: -1, RandomTickSpeed: -1}.New()

	errCh := make(chan error, 1)
	done := w.Exec(func(tx *Tx) {
		pos := ChunkPos{}
		_ = tx.chunk(pos)

		l := NewLoader(1, tx.World(), NopViewer{})
		l.Load(tx, 1)

		if _, ok := l.Chunk(pos); !ok {
			errCh <- errors.New("expected chunk to be marked loaded")
		}
	})

	select {
	case <-done:
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Load deadlocked when the chunk callback ran synchronously")
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func TestLoaderViewChunkSerializesLoadedAccess(t *testing.T) {
	w := Config{SaveInterval: -1, RandomTickSpeed: -1}.New()

	errCh := make(chan error, 1)
	done := w.Exec(func(tx *Tx) {
		pos := ChunkPos{}
		c := tx.chunk(pos)
		l := NewLoader(1, tx.World(), NopViewer{})

		stop := make(chan struct{})
		readerDone := make(chan struct{})
		go func() {
			defer close(readerDone)
			for {
				select {
				case <-stop:
					return
				default:
					l.Chunk(pos)
				}
			}
		}()

		for i := 0; i < 1_000; i++ {
			l.viewChunk(tx, pos, c)
		}
		close(stop)
		<-readerDone

		if _, ok := l.Chunk(pos); !ok {
			errCh <- errors.New("expected chunk to remain loaded")
		}
	})

	select {
	case <-done:
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("viewChunk did not complete while Chunk was reading")
	}

	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func TestAsyncChunkLoadsAreBounded(t *testing.T) {
	const workers = 2
	provider := &blockingProvider{
		release: make(chan struct{}),
		limit:   workers,
		max:     make(chan struct{}, workers+1),
	}
	w := Config{Provider: provider, SaveInterval: -1, ChunkLoadWorkers: workers, RandomTickSpeed: -1}.New()

	done := w.Exec(func(tx *Tx) {
		for x := range workers + 1 {
			pos := ChunkPos{int32(x), 0}
			tx.World().loadChunkAsync(tx, pos, func(*Tx, *Column) {})
		}
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("queueing chunk loads blocked the transaction")
	}

	select {
	case <-provider.max:
		t.Fatal("started more concurrent chunk loads than the worker limit")
	case <-time.After(100 * time.Millisecond):
	}

	close(provider.release)
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

type blockingProvider struct {
	NopProvider
	release chan struct{}
	active  atomic.Int32
	limit   int32
	max     chan struct{}
}

func (p *blockingProvider) LoadColumn(ChunkPos, Dimension) (*chunk.Column, error) {
	active := p.active.Add(1)
	defer p.active.Add(-1)

	if active > p.limit {
		p.max <- struct{}{}
	}
	<-p.release
	return nil, leveldb.ErrNotFound
}

func (p *blockingProvider) Settings() *Settings                                  { return defaultSettings() }
func (p *blockingProvider) SaveSettings(*Settings)                               {}
func (p *blockingProvider) StoreColumn(ChunkPos, Dimension, *chunk.Column) error { return nil }
func (p *blockingProvider) LoadPlayerSpawnPosition(uuid.UUID) (cube.Pos, bool, error) {
	return cube.Pos{}, false, nil
}
func (p *blockingProvider) SavePlayerSpawnPosition(uuid.UUID, cube.Pos) error { return nil }
func (p *blockingProvider) Close() error                                      { return nil }
