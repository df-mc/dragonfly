package world

import (
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
)

func trackerWorld() *World {
	return &World{set: defaultSettings(), ra: cube.Range{-64, 319}}
}

// TestTrackPositionReusesActiveHandle verifies that linking twice to a standing block keeps one handle.
func TestTrackPositionReusesActiveHandle(t *testing.T) {
	w, pos := trackerWorld(), cube.Pos{0, 0, 0}
	if first, second := w.TrackPosition(pos, 0), w.TrackPosition(pos, 0); first != second {
		t.Errorf("a standing block should keep handle %v, got %v", first, second)
	}
}

// TestUntrackPositionFormsNewGroup verifies that a block rebuilt in the same spot forms a new group, leaving
// compasses linked to the old one unresolved.
func TestUntrackPositionFormsNewGroup(t *testing.T) {
	w, pos := trackerWorld(), cube.Pos{0, 0, 0}

	first := w.TrackPosition(pos, 0)
	w.UntrackPosition(pos)
	if _, _, ok := w.TrackedPosition(first); ok {
		t.Error("a retired handle should stop resolving")
	}
	if second := w.TrackPosition(pos, 0); second == first {
		t.Errorf("a rebuilt block reused handle %v instead of forming a new group", first)
	}
}

// TestUntrackPositionDropsEntry verifies that retiring a handle leaves nothing behind to persist.
func TestUntrackPositionDropsEntry(t *testing.T) {
	w, pos := trackerWorld(), cube.Pos{1, 2, 3}

	w.TrackPosition(pos, 0)
	w.UntrackPosition(pos)
	if n := len(w.set.PositionTrackingData().Entries); n != 0 {
		t.Errorf("expected no persisted entries after retiring, got %v", n)
	}
	if handle := w.PositionTrackingHandleAt(pos); handle != 0 {
		t.Errorf("expected no handle at a retired position, got %v", handle)
	}
}

// TestAdoptedHandleRaisesNext verifies that a handle read back from a saved block cannot be handed out again,
// which is what happens in worlds carrying no tracking data of their own.
func TestAdoptedHandleRaisesNext(t *testing.T) {
	w := trackerWorld()

	if adopted := w.TrackPosition(cube.Pos{0, 0, 0}, 100); adopted != 100 {
		t.Fatalf("expected the saved handle to be kept, got %v", adopted)
	}
	if fresh := w.TrackPosition(cube.Pos{9, 9, 9}, 0); fresh <= 100 {
		t.Errorf("handle %v collides with the adopted one", fresh)
	}
}

// TestPositionTrackingRoundTrip verifies that entries and the next handle survive a save and load.
func TestPositionTrackingRoundTrip(t *testing.T) {
	w, pos := trackerWorld(), cube.Pos{4, 5, 6}
	handle := w.TrackPosition(pos, 0)

	loaded := trackerWorld()
	loaded.set.LoadPositionTrackingData(w.set.PositionTrackingData())

	if p, _, ok := loaded.TrackedPosition(handle); !ok || p != pos {
		t.Errorf("handle %v resolved to %v (%v) after a round trip, want %v", handle, p, ok, pos)
	}
	if next := loaded.TrackPosition(cube.Pos{7, 8, 9}, 0); next == handle {
		t.Errorf("handle %v was handed out twice across a round trip", handle)
	}
}

// TestTrackerReachableWhileSettingsLocked covers the world tick, which holds the Settings lock while the
// spawn height calculation loads chunks, and chunk loading reaches the tracker.
func TestTrackerReachableWhileSettingsLocked(t *testing.T) {
	w := trackerWorld()

	done := make(chan struct{})
	go func() {
		w.set.Lock()
		w.TrackPosition(cube.Pos{0, 0, 0}, 0)
		w.set.Unlock()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reaching the tracker while the Settings lock is held deadlocked")
	}
}
