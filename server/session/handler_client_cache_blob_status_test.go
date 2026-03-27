package session

import (
	"testing"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestClientCacheBlobStatusHandlerObservesDeferredChunkVisibility(t *testing.T) {
	w := world.Config{Provider: world.NopProvider{}, SaveInterval: -1}.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()

	s := &Session{
		conf:                         Config{MetricsLogThreshold: -1},
		packets:                      make(chan packet.Packet, 1),
		closeBackground:              make(chan struct{}),
		blobs:                        map[uint64][]byte{15: {1, 2, 3}},
		openChunkTransactions:        []map[uint64]struct{}{{11: {}, 12: {}, 13: {}, 14: {}, 15: {}}},
		pendingChunkVisibilityByBlob: map[uint64]map[world.ChunkPos]struct{}{},
	}
	s.chunkMetrics = newChunkVisibilityTracker(w, mgl64.Vec3{8, 64, 8}, "join")

	required := immediateChunkPositions(world.ChunkPos{0, 0})
	for i, pos := range required {
		s.deferChunkVisibilityForBlob(uint64(11+i), pos)
	}

	handler := &ClientCacheBlobStatusHandler{}
	if err := handler.Handle(&packet.ClientCacheBlobStatus{
		HitHashes:  []uint64{11, 12, 13, 14},
		MissHashes: []uint64{15},
	}, s, nil, nil); err != nil {
		t.Fatalf("handle blob status: %v", err)
	}

	if !s.chunkMetrics.centerObserved {
		t.Fatal("expected deferred center chunk visibility to resolve after blob status")
	}
	if !s.chunkMetrics.immediateVisible {
		t.Fatal("expected deferred immediate-neighbor visibility to resolve after blob status")
	}

	select {
	case pk := <-s.packets:
		if _, ok := pk.(*packet.ClientCacheMissResponse); !ok {
			t.Fatalf("expected miss response packet, got %T", pk)
		}
	default:
		t.Fatal("expected miss response packet to be queued")
	}
}

func TestClientCacheBlobStatusHandlerSkipsDeferredVisibilityWhenBlobMissing(t *testing.T) {
	w := world.Config{Provider: world.NopProvider{}, SaveInterval: -1}.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()

	s := &Session{
		conf:                         Config{MetricsLogThreshold: -1},
		packets:                      make(chan packet.Packet, 1),
		closeBackground:              make(chan struct{}),
		blobs:                        map[uint64][]byte{},
		openChunkTransactions:        []map[uint64]struct{}{{21: {}}},
		pendingChunkVisibilityByBlob: map[uint64]map[world.ChunkPos]struct{}{},
	}
	s.chunkMetrics = newChunkVisibilityTracker(w, mgl64.Vec3{8, 64, 8}, "join")
	s.deferChunkVisibilityForBlob(21, world.ChunkPos{0, 0})

	handler := &ClientCacheBlobStatusHandler{}
	if err := handler.Handle(&packet.ClientCacheBlobStatus{
		MissHashes: []uint64{21},
	}, s, nil, nil); err != nil {
		t.Fatalf("handle blob status: %v", err)
	}

	if s.chunkMetrics.centerObserved {
		t.Fatal("expected missing blobs to leave deferred chunk visibility unresolved")
	}

	select {
	case pk := <-s.packets:
		t.Fatalf("expected no packet to be queued when blobs are missing, got %T", pk)
	default:
	}
}
