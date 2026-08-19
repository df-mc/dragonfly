package world

import (
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// PositionTrackingBlock is implemented by blocks whose position may be tracked
// by the Bedrock client, currently only lodestones.
type PositionTrackingBlock interface {
	Block
	TrackingHandle() int32
	WithTrackingHandle(handle int32) Block
}

type trackedPosition struct {
	pos cube.Pos
	dim int
}

// PositionTrackingEntry is a persistent entry in the position tracking database.
type PositionTrackingEntry struct {
	Handle    int32
	Position  cube.Pos
	Dimension int
}

// PositionTrackingDestroyAction notifies viewers that a tracked block no
// longer exists, causing matching compasses to spin immediately.
type PositionTrackingDestroyAction struct {
	Handle int32
}

// BlockAction serves to implement the world.BlockAction interface.
func (PositionTrackingDestroyAction) BlockAction() {}

// PositionTrackingUpdateAction provides the target of a tracking handle to a
// viewer immediately after a compass is linked.
type PositionTrackingUpdateAction struct {
	Handle    int32
	Position  cube.Pos
	Dimension int
}

// BlockAction serves to implement the world.BlockAction interface.
func (PositionTrackingUpdateAction) BlockAction() {}

// PositionTracker holds Bedrock position tracking handles shared by the dimensions of a world.
type PositionTracker struct {
	mu         sync.Mutex
	next       int32
	byHandle   map[int32]trackedPosition
	byPosition map[[4]int]int32
}

func newPositionTracker() *PositionTracker {
	return &PositionTracker{byHandle: map[int32]trackedPosition{}, byPosition: map[[4]int]int32{}}
}

// tracker returns the PositionTracker of the Settings, creating it if absent. It deliberately avoids the
// Settings mutex: chunk loading reaches this while the world tick already holds that lock.
func (s *Settings) tracker() *PositionTracker {
	s.trackerMu.Lock()
	defer s.trackerMu.Unlock()
	if s.positionTracker == nil {
		s.positionTracker = newPositionTracker()
	}
	return s.positionTracker
}

// PositionTrackingData is a persistent snapshot of the position tracking database.
type PositionTrackingData struct {
	Next    int32
	Entries []PositionTrackingEntry
}

// PositionTrackingData returns a snapshot of the position tracking database.
func (s *Settings) PositionTrackingData() PositionTrackingData {
	t := s.tracker()
	t.mu.Lock()
	defer t.mu.Unlock()
	data := PositionTrackingData{Next: t.next, Entries: make([]PositionTrackingEntry, 0, len(t.byHandle))}
	for handle, entry := range t.byHandle {
		data.Entries = append(data.Entries, PositionTrackingEntry{Handle: handle, Position: entry.pos, Dimension: entry.dim})
	}
	return data
}

// LoadPositionTrackingData replaces the position tracking database with data.
func (s *Settings) LoadPositionTrackingData(data PositionTrackingData) {
	t := s.tracker()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.next = data.Next
	t.byHandle = map[int32]trackedPosition{}
	t.byPosition = map[[4]int]int32{}
	for _, entry := range data.Entries {
		if entry.Handle == 0 {
			continue
		}
		t.byHandle[entry.Handle] = trackedPosition{pos: entry.Position, dim: entry.Dimension}
		t.byPosition[[4]int{entry.Dimension, entry.Position[0], entry.Position[1], entry.Position[2]}] = entry.Handle
		if entry.Handle > t.next {
			t.next = entry.Handle
		}
	}
}

// retrackBlock moves the position tracking database from the block previously at pos to b, returning the
// block to store. Every path that replaces blocks must call it, not just Tx.setBlock: a structure dropping a
// lodestone would otherwise leave its handle resolving to a block that no longer exists.
func (w *World) retrackBlock(pos cube.Pos, old, b Block) Block {
	if tracked, ok := old.(PositionTrackingBlock); ok && tracked.TrackingHandle() != 0 {
		replacement, keepsHandle := b.(PositionTrackingBlock)
		if !keepsHandle || replacement.TrackingHandle() != tracked.TrackingHandle() {
			w.UntrackPosition(pos)
		}
	}
	if tracked, ok := b.(PositionTrackingBlock); ok {
		if handle := tracked.TrackingHandle(); handle != 0 || w.PositionTrackingHandleAt(pos) != 0 {
			return tracked.WithTrackingHandle(w.TrackPosition(pos, handle))
		}
	}
	return b
}

// TrackPosition assigns a tracking handle to pos, reusing the one already there if any.
func (w *World) TrackPosition(pos cube.Pos, handle int32) int32 {
	dim, _ := DimensionID(w.Dimension())
	t := w.set.tracker()
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.byHandle == nil {
		t.byHandle = map[int32]trackedPosition{}
		t.byPosition = map[[4]int]int32{}
	}
	key := [4]int{dim, pos[0], pos[1], pos[2]}
	// The wiki states a lodestone rebuilt in the same spot keeps compasses linked to the old one, but in-game
	// testing shows it forms a new group, so UntrackPosition drops the handle rather than parking it.
	if existing := t.byPosition[key]; existing != 0 {
		handle = existing
	}
	if handle == 0 {
		t.next++
		handle = t.next
	} else if handle > t.next {
		// A handle adopted from a saved block must raise the counter, or a later lodestone is issued the same
		// one and steals the compasses linked to it.
		t.next = handle
	}
	if entry, ok := t.byHandle[handle]; ok {
		delete(t.byPosition, [4]int{entry.dim, entry.pos[0], entry.pos[1], entry.pos[2]})
	}
	t.byPosition[key] = handle
	t.byHandle[handle] = trackedPosition{pos: pos, dim: dim}
	return handle
}

// PositionTrackingHandleAt returns the tracking handle at pos, or 0 if there is none.
func (w *World) PositionTrackingHandleAt(pos cube.Pos) int32 {
	dim, _ := DimensionID(w.Dimension())
	t := w.set.tracker()
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.byPosition[[4]int{dim, pos[0], pos[1], pos[2]}]
}

// UntrackPosition retires the tracking handle at pos, making compasses linked to it spin.
func (w *World) UntrackPosition(pos cube.Pos) {
	dim, _ := DimensionID(w.Dimension())
	t := w.set.tracker()
	t.mu.Lock()
	key := [4]int{dim, pos[0], pos[1], pos[2]}
	handle := t.byPosition[key]
	delete(t.byPosition, key)
	delete(t.byHandle, handle)
	t.mu.Unlock()
	if handle == 0 {
		return
	}
	action := PositionTrackingDestroyAction{Handle: handle}
	w.viewerMu.Lock()
	viewers := make(map[Viewer]struct{}, len(w.viewers))
	for _, viewer := range w.viewers {
		viewers[viewer] = struct{}{}
	}
	w.viewerMu.Unlock()
	for viewer := range viewers {
		viewer.ViewBlockAction(pos, action)
	}
}

// TrackedPosition returns the position a tracking handle resolves to, if it still does.
func (w *World) TrackedPosition(handle int32) (cube.Pos, int, bool) {
	t := w.set.tracker()
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.byHandle[handle]
	if !ok {
		return cube.Pos{}, 0, false
	}
	return entry.pos, entry.dim, true
}
