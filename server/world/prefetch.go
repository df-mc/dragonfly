package world

import (
	"time"

	"github.com/df-mc/dragonfly/server/world/chunk"
)

const prefetchWorkers = 4

func (w *World) prefetchLoop() {
	defer w.running.Done()

	for {
		select {
		case <-w.closing:
			return
		case pos := <-w.prefetchRequests:
			w.prefetchOne(pos)
		}
	}
}

func (w *World) prefetchOne(pos ChunkPos) {
	loaded := w.loadColumnData(pos, "prefetch")
	if loaded.column == nil {
		loaded.column = &chunk.Column{Chunk: chunk.New(airRID, w.Range())}
	}

	fillStart := time.Now()
	chunk.LightArea([]*chunk.Chunk{loaded.column.Chunk}, int(pos[0]), int(pos[1])).Fill()
	loaded.fillTime = time.Since(fillStart)
	w.observeLighting(pos, "prefetch", loaded.fillTime)

	select {
	case <-w.closing:
		w.clearPrefetchInFlight(pos)
		return
	default:
	}

	_ = w.Exec(func(tx *Tx) {
		tx.World().installPrefetched(pos, loaded)
	})
}

func (w *World) installPrefetched(pos ChunkPos, loaded loadedColumn) {
	defer w.clearPrefetchInFlight(pos)

	if _, ok := w.chunks[pos]; ok {
		return
	}

	start := time.Now()
	col := w.columnFrom(loaded.column, pos)
	col.modified = loaded.generated
	w.chunks[pos] = col
	for _, e := range col.Entities {
		w.entities[e] = pos
		e.w = w
	}
	w.calculateLight(pos)
	w.observeInstallation(pos, "prefetch", time.Since(start))

	if loaded.err != nil {
		w.conf.Log.Error("load chunk: "+loaded.err.Error(), "X", pos[0], "Z", pos[1])
	}
}

func (w *World) requestPrefetch(pos ChunkPos) bool {
	if _, ok := w.chunks[pos]; ok {
		return false
	}

	w.prefetchMu.Lock()
	if _, ok := w.prefetchInFlight[pos]; ok {
		w.prefetchMu.Unlock()
		return false
	}
	w.prefetchInFlight[pos] = struct{}{}
	w.prefetchMu.Unlock()

	select {
	case <-w.closing:
		w.clearPrefetchInFlight(pos)
		return false
	case w.prefetchRequests <- pos:
		w.metrics.prefetchQueued.Add(1)
		return true
	default:
		w.metrics.prefetchDropped.Add(1)
		w.clearPrefetchInFlight(pos)
		return false
	}
}

func (w *World) clearPrefetchInFlight(pos ChunkPos) {
	w.prefetchMu.Lock()
	delete(w.prefetchInFlight, pos)
	w.prefetchMu.Unlock()
}
