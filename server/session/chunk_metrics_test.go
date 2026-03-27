package session

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestChunkVisibilityTrackerEmitsCenterAndImmediateStages(t *testing.T) {
	w := world.Config{Provider: world.NopProvider{}, SaveInterval: -1}.New()
	defer func() {
		if err := w.Close(); err != nil {
			t.Fatalf("close world: %v", err)
		}
	}()

	tracker := newChunkVisibilityTracker(w, mgl64.Vec3{8, 64, 8}, "join")
	if events := tracker.observe(world.ChunkPos{3, 3}, tracker.startedAt.Add(10*time.Millisecond)); len(events) != 0 {
		t.Fatalf("expected unrelated chunk to be ignored, got %v", events)
	}

	center := world.ChunkPos{0, 0}
	events := tracker.observe(center, tracker.startedAt.Add(50*time.Millisecond))
	if len(events) != 1 || events[0].stage != "center" {
		t.Fatalf("expected center stage event, got %+v", events)
	}
	if events := tracker.observe(center, tracker.startedAt.Add(60*time.Millisecond)); len(events) != 0 {
		t.Fatalf("expected duplicate center observation to be ignored, got %v", events)
	}

	required := immediateChunkPositions(center)
	for i, pos := range required[1:] {
		events = tracker.observe(pos, tracker.startedAt.Add(time.Duration(100+i)*time.Millisecond))
		if i < len(required)-2 && len(events) != 0 {
			t.Fatalf("expected no immediate-neighbor event before final chunk, got %+v", events)
		}
	}
	if len(events) != 1 || events[0].stage != "immediate_neighbors" {
		t.Fatalf("expected immediate-neighbor stage event, got %+v", events)
	}
}

func TestChunkPosFromVec3FloorsNegativeCoordinates(t *testing.T) {
	pos := chunkPosFromVec3(mgl64.Vec3{-0.1, 70, -16.1})
	if pos != (world.ChunkPos{-1, -2}) {
		t.Fatalf("expected floored negative chunk position (-1, -2), got %v", pos)
	}
}
