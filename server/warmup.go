package server

import (
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	defaultSpawnWarmupRadius = 2
	spawnWarmupPollInterval  = 10 * time.Millisecond
	spawnWarmupTimeout       = 15 * time.Second
)

func (srv *Server) warmupSpawnArea(w *world.World) *world.Loader {
	if w == nil || w.Dimension() != world.Overworld {
		return nil
	}

	radius := srv.conf.SpawnWarmupRadius
	switch {
	case radius < 0:
		return nil
	case radius == 0:
		radius = defaultSpawnWarmupRadius
	}

	spawn := w.Spawn()
	center := world.ChunkPos{int32(spawn[0] >> 4), int32(spawn[2] >> 4)}
	required := warmupChunkPositions(center, radius)
	loader := world.NewLoader(radius, w, world.NopViewer{})
	target := mgl64.Vec3{float64(spawn[0]) + 0.5, float64(spawn[1]), float64(spawn[2]) + 0.5}

	start := time.Now()
	baseline := w.Metrics()
	if !waitForWarmupChunks(w, loader, target, required, spawnWarmupTimeout) {
		srv.conf.Log.Warn("Spawn warmup timed out.",
			append([]any{
				"spawn", spawn,
				"chunk", center,
				"radius", radius,
				"required_chunks", len(required),
				"loaded_chunks", countWarmupChunks(loader, required),
				"duration", time.Since(start),
			}, warmupMetricsAttrs(w.Metrics().DeltaSince(baseline))...)...,
		)
		return loader
	}

	srv.conf.Log.Info("Spawn warmup completed.",
		append([]any{
			"spawn", spawn,
			"chunk", center,
			"radius", radius,
			"chunks", len(required),
			"duration", time.Since(start),
		}, warmupMetricsAttrs(w.Metrics().DeltaSince(baseline))...)...,
	)
	return loader
}

func waitForWarmupChunks(w *world.World, loader *world.Loader, pos mgl64.Vec3, required []world.ChunkPos, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		<-w.Exec(func(tx *world.Tx) {
			loader.Move(tx, pos)
			loader.Load(tx, len(required))
		})
		if countWarmupChunks(loader, required) == len(required) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(spawnWarmupPollInterval)
	}
}

func countWarmupChunks(loader *world.Loader, required []world.ChunkPos) int {
	count := 0
	for _, pos := range required {
		if _, ok := loader.Chunk(pos); ok {
			count++
		}
	}
	return count
}

func warmupChunkPositions(center world.ChunkPos, radius int) []world.ChunkPos {
	if radius < 0 {
		return nil
	}

	out := make([]world.ChunkPos, 0, (2*radius+1)*(2*radius+1))
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			if x*x+z*z > radius*radius {
				continue
			}
			out = append(out, world.ChunkPos{center[0] + int32(x), center[1] + int32(z)})
		}
	}
	return out
}

func warmupMetricsAttrs(metrics world.MetricsSnapshot) []any {
	return []any{
		"provider_hits", metrics.ProviderHits,
		"provider_misses", metrics.ProviderMisses,
		"provider_errors", metrics.ProviderErrors,
		"provider_load_avg", metrics.ProviderLoad.Average,
		"generated_chunks", metrics.Generation.Count,
		"generation_avg", metrics.Generation.Average,
		"lit_chunks", metrics.Lighting.Count,
		"lighting_avg", metrics.Lighting.Average,
		"installed_chunks", metrics.Installation.Count,
		"installation_avg", metrics.Installation.Average,
		"sync_chunks", metrics.SyncChunk.Count,
		"sync_avg", metrics.SyncChunk.Average,
		"prefetch_queued", metrics.PrefetchQueuedTotal,
		"prefetch_dropped", metrics.PrefetchDropped,
		"prefetch_queue_depth", metrics.PrefetchQueueDepth,
		"prefetch_in_flight", metrics.PrefetchInFlight,
	}
}
