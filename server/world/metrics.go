package world

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/df-mc/goleveldb/leveldb"
)

const defaultMetricsLogThreshold = 250 * time.Millisecond

type durationMetric struct {
	count      atomic.Int64
	totalNanos atomic.Int64
}

func (m *durationMetric) Observe(d time.Duration) {
	m.count.Add(1)
	m.totalNanos.Add(d.Nanoseconds())
}

func (m *durationMetric) Snapshot() DurationMetricSnapshot {
	count := m.count.Load()
	total := time.Duration(m.totalNanos.Load())

	var average time.Duration
	if count > 0 {
		average = total / time.Duration(count)
	}
	return DurationMetricSnapshot{
		Count:   count,
		Total:   total,
		Average: average,
	}
}

type worldMetrics struct {
	providerLoad durationMetric
	generation   durationMetric
	lighting     durationMetric
	installation durationMetric
	syncChunk    durationMetric

	providerHits   atomic.Int64
	providerMisses atomic.Int64
	providerErrors atomic.Int64

	prefetchQueued  atomic.Int64
	prefetchDropped atomic.Int64
}

// DurationMetricSnapshot captures aggregate timing data for a repeated
// operation.
type DurationMetricSnapshot struct {
	Count   int64
	Total   time.Duration
	Average time.Duration
}

// MetricsSnapshot captures the current runtime chunk-loading metrics of a
// world.
type MetricsSnapshot struct {
	ProviderHits   int64
	ProviderMisses int64
	ProviderErrors int64

	ProviderLoad DurationMetricSnapshot
	Generation   DurationMetricSnapshot
	Lighting     DurationMetricSnapshot
	Installation DurationMetricSnapshot
	SyncChunk    DurationMetricSnapshot

	PrefetchQueuedTotal int64
	PrefetchDropped     int64
	PrefetchQueueDepth  int
	PrefetchInFlight    int
}

// Metrics returns a snapshot of the runtime chunk-loading metrics of the
// world.
func (w *World) Metrics() MetricsSnapshot {
	w.prefetchMu.Lock()
	inFlight := len(w.prefetchInFlight)
	w.prefetchMu.Unlock()

	return MetricsSnapshot{
		ProviderHits:        w.metrics.providerHits.Load(),
		ProviderMisses:      w.metrics.providerMisses.Load(),
		ProviderErrors:      w.metrics.providerErrors.Load(),
		ProviderLoad:        w.metrics.providerLoad.Snapshot(),
		Generation:          w.metrics.generation.Snapshot(),
		Lighting:            w.metrics.lighting.Snapshot(),
		Installation:        w.metrics.installation.Snapshot(),
		SyncChunk:           w.metrics.syncChunk.Snapshot(),
		PrefetchQueuedTotal: w.metrics.prefetchQueued.Load(),
		PrefetchDropped:     w.metrics.prefetchDropped.Load(),
		PrefetchQueueDepth:  len(w.prefetchRequests),
		PrefetchInFlight:    inFlight,
	}
}

// DeltaSince returns the cumulative metrics observed since the earlier
// snapshot. Instantaneous queue depth and in-flight counts reflect the current
// snapshot.
func (m MetricsSnapshot) DeltaSince(earlier MetricsSnapshot) MetricsSnapshot {
	return MetricsSnapshot{
		ProviderHits:        m.ProviderHits - earlier.ProviderHits,
		ProviderMisses:      m.ProviderMisses - earlier.ProviderMisses,
		ProviderErrors:      m.ProviderErrors - earlier.ProviderErrors,
		ProviderLoad:        m.ProviderLoad.deltaSince(earlier.ProviderLoad),
		Generation:          m.Generation.deltaSince(earlier.Generation),
		Lighting:            m.Lighting.deltaSince(earlier.Lighting),
		Installation:        m.Installation.deltaSince(earlier.Installation),
		SyncChunk:           m.SyncChunk.deltaSince(earlier.SyncChunk),
		PrefetchQueuedTotal: m.PrefetchQueuedTotal - earlier.PrefetchQueuedTotal,
		PrefetchDropped:     m.PrefetchDropped - earlier.PrefetchDropped,
		PrefetchQueueDepth:  m.PrefetchQueueDepth,
		PrefetchInFlight:    m.PrefetchInFlight,
	}
}

func (w *World) metricsLogThreshold() time.Duration {
	switch {
	case w.conf.MetricsLogThreshold < 0:
		return -1
	case w.conf.MetricsLogThreshold == 0:
		return defaultMetricsLogThreshold
	default:
		return w.conf.MetricsLogThreshold
	}
}

func (w *World) logSlowMetric(message string, duration time.Duration, attrs ...any) {
	threshold := w.metricsLogThreshold()
	if threshold < 0 || duration < threshold {
		return
	}
	attrs = append(attrs, "duration", duration)
	w.conf.Log.Info(message, attrs...)
}

func (w *World) observeProviderLoad(pos ChunkPos, source string, duration time.Duration, err error) {
	w.metrics.providerLoad.Observe(duration)

	result := "hit"
	switch {
	case err == nil:
		w.metrics.providerHits.Add(1)
	case errors.Is(err, leveldb.ErrNotFound):
		result = "miss"
		w.metrics.providerMisses.Add(1)
	default:
		result = "error"
		w.metrics.providerErrors.Add(1)
	}
	w.logSlowMetric("chunk provider load slow", duration, "chunk", pos, "source", source, "result", result)
}

func (w *World) observeGeneration(pos ChunkPos, source string, duration time.Duration) {
	w.metrics.generation.Observe(duration)
	w.logSlowMetric("chunk generation slow", duration, "chunk", pos, "source", source)
}

func (w *World) observeLighting(pos ChunkPos, source string, duration time.Duration) {
	w.metrics.lighting.Observe(duration)
	w.logSlowMetric("chunk lighting slow", duration, "chunk", pos, "source", source)
}

func (w *World) observeInstallation(pos ChunkPos, source string, duration time.Duration) {
	w.metrics.installation.Observe(duration)
	w.logSlowMetric("chunk installation slow", duration, "chunk", pos, "source", source)
}

func (w *World) observeSyncChunkLoad(pos ChunkPos, duration time.Duration) {
	w.metrics.syncChunk.Observe(duration)
	w.logSlowMetric("synchronous chunk load slow", duration, "chunk", pos)
}

func (m DurationMetricSnapshot) deltaSince(earlier DurationMetricSnapshot) DurationMetricSnapshot {
	count := m.Count - earlier.Count
	total := m.Total - earlier.Total

	var average time.Duration
	if count > 0 {
		average = total / time.Duration(count)
	}
	return DurationMetricSnapshot{
		Count:   count,
		Total:   total,
		Average: average,
	}
}
