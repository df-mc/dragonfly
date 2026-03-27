package session

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

const defaultMetricsLogThreshold = 250 * time.Millisecond

type chunkVisibilityTracker struct {
	world     *world.World
	reason    string
	startedAt time.Time
	center    world.ChunkPos
	baseline  world.MetricsSnapshot

	seen             map[world.ChunkPos]struct{}
	centerObserved   bool
	immediateVisible bool
}

type chunkVisibilityEvent struct {
	reason   string
	stage    string
	duration time.Duration
	center   world.ChunkPos
	metrics  world.MetricsSnapshot
}

func newChunkVisibilityTracker(w *world.World, pos mgl64.Vec3, reason string) chunkVisibilityTracker {
	var baseline world.MetricsSnapshot
	if w != nil {
		baseline = w.Metrics()
	}
	return chunkVisibilityTracker{
		world:     w,
		reason:    reason,
		startedAt: time.Now(),
		center:    chunkPosFromVec3(pos),
		baseline:  baseline,
		seen:      make(map[world.ChunkPos]struct{}, 5),
	}
}

func (t *chunkVisibilityTracker) observe(pos world.ChunkPos, now time.Time) []chunkVisibilityEvent {
	if t.world == nil || t.startedAt.IsZero() || !t.tracks(pos) {
		return nil
	}
	if _, ok := t.seen[pos]; ok {
		return nil
	}
	t.seen[pos] = struct{}{}

	events := make([]chunkVisibilityEvent, 0, 2)
	if !t.centerObserved && pos == t.center {
		t.centerObserved = true
		events = append(events, t.event(now, "center"))
	}
	if !t.immediateVisible && len(t.seen) == 5 {
		t.immediateVisible = true
		events = append(events, t.event(now, "immediate_neighbors"))
	}
	return events
}

func (t *chunkVisibilityTracker) event(now time.Time, stage string) chunkVisibilityEvent {
	return chunkVisibilityEvent{
		reason:   t.reason,
		stage:    stage,
		duration: now.Sub(t.startedAt),
		center:   t.center,
		metrics:  t.world.Metrics().DeltaSince(t.baseline),
	}
}

func (t *chunkVisibilityTracker) tracks(pos world.ChunkPos) bool {
	if pos == t.center {
		return true
	}
	required := immediateChunkPositions(t.center)
	for _, chunkPos := range required[1:] {
		if pos == chunkPos {
			return true
		}
	}
	return false
}

func immediateChunkPositions(center world.ChunkPos) [5]world.ChunkPos {
	return [5]world.ChunkPos{
		center,
		{center[0] - 1, center[1]},
		{center[0], center[1] - 1},
		{center[0], center[1] + 1},
		{center[0] + 1, center[1]},
	}
}

func chunkPosFromVec3(pos mgl64.Vec3) world.ChunkPos {
	return world.ChunkPos{
		int32(math.Floor(pos[0])) >> 4,
		int32(math.Floor(pos[2])) >> 4,
	}
}

func (s *Session) beginChunkVisibilityTracking(w *world.World, pos mgl64.Vec3, reason string) {
	s.chunkMetricsMu.Lock()
	s.chunkMetrics = newChunkVisibilityTracker(w, pos, reason)
	s.chunkMetricsMu.Unlock()
	s.clearDeferredChunkVisibility()
}

func (s *Session) clearChunkVisibilityTracking() {
	s.chunkMetricsMu.Lock()
	s.chunkMetrics = chunkVisibilityTracker{}
	s.chunkMetricsMu.Unlock()
	s.clearDeferredChunkVisibility()
}

func (s *Session) observeChunkVisibility(pos world.ChunkPos) {
	s.chunkMetricsMu.Lock()
	events := s.chunkMetrics.observe(pos, time.Now())
	s.chunkMetricsMu.Unlock()

	for _, event := range events {
		s.logChunkVisibilityEvent(event)
	}
}

func (s *Session) logChunkVisibilityEvent(event chunkVisibilityEvent) {
	threshold := s.metricsLogThreshold()
	if threshold < 0 {
		return
	}

	attrs := []any{
		"reason", event.reason,
		"stage", event.stage,
		"center_chunk", event.center,
		"duration", event.duration,
		"provider_hits", event.metrics.ProviderHits,
		"provider_misses", event.metrics.ProviderMisses,
		"provider_errors", event.metrics.ProviderErrors,
		"provider_loads", event.metrics.ProviderLoad.Count,
		"provider_load_avg", event.metrics.ProviderLoad.Average,
		"generated_chunks", event.metrics.Generation.Count,
		"generation_avg", event.metrics.Generation.Average,
		"lit_chunks", event.metrics.Lighting.Count,
		"lighting_avg", event.metrics.Lighting.Average,
		"installed_chunks", event.metrics.Installation.Count,
		"installation_avg", event.metrics.Installation.Average,
		"sync_chunks", event.metrics.SyncChunk.Count,
		"sync_avg", event.metrics.SyncChunk.Average,
		"prefetch_queued", event.metrics.PrefetchQueuedTotal,
		"prefetch_dropped", event.metrics.PrefetchDropped,
		"prefetch_queue_depth", event.metrics.PrefetchQueueDepth,
		"prefetch_in_flight", event.metrics.PrefetchInFlight,
	}
	if event.duration >= threshold {
		s.conf.Log.Info("chunk stream milestone", attrs...)
		return
	}
	s.conf.Log.Debug("chunk stream milestone", attrs...)
}

func (s *Session) metricsLogThreshold() time.Duration {
	switch {
	case s.conf.MetricsLogThreshold < 0:
		return -1
	case s.conf.MetricsLogThreshold == 0:
		return defaultMetricsLogThreshold
	default:
		return s.conf.MetricsLogThreshold
	}
}

func (s *Session) clearDeferredChunkVisibility() {
	s.blobMu.Lock()
	s.pendingChunkVisibilityByBlob = map[uint64]map[world.ChunkPos]struct{}{}
	s.blobMu.Unlock()
}
