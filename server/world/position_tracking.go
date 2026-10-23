package world

import (
	"sync"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// PositionTrackingBlock is a block whose position compasses may track, such as a lodestone.
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

// PositionTrackingDestroyAction tells viewers a tracked block is gone, making compasses linked to it spin.
type PositionTrackingDestroyAction struct {
	Handle int32
}

// BlockAction serves to implement the world.BlockAction interface.
func (PositionTrackingDestroyAction) BlockAction() {}

// PositionTrackingUpdateAction sends viewers the target of a tracking handle.
type PositionTrackingUpdateAction struct {
	Handle    int32
	Position  cube.Pos
	Dimension int
}

// BlockAction serves to implement the world.BlockAction interface.
func (PositionTrackingUpdateAction) BlockAction() {}

// PositionTrackingStatus is the state of a tracking handle.
type PositionTrackingStatus uint8

const (
	// PositionNotFound is the status of a handle that was never issued, or was destroyed before the last restart.
	PositionNotFound PositionTrackingStatus = iota
	// PositionTracked is the status of a handle that resolves to a block.
	PositionTracked
	// PositionDestroyed is the status of a handle whose block was destroyed since the last restart.
	PositionDestroyed
)

// PositionTracker holds Bedrock position tracking handles shared by the dimensions of a world.
type PositionTracker struct {
	mu         sync.Mutex
	next       int32
	byHandle   map[int32]trackedPosition
	byPosition map[[4]int]int32
	// destroyed is not saved: vanilla forgets destroyed handles on restart.
	destroyed map[int32]struct{}
}

func newPositionTracker() *PositionTracker {
	return &PositionTracker{byHandle: map[int32]trackedPosition{}, byPosition: map[[4]int]int32{}, destroyed: map[int32]struct{}{}}
}

// tracker returns the PositionTracker of the Settings. It avoids the Settings mutex, which chunk loading holds.
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

// retrackBlock moves the tracking handle from the block previously at pos to b and returns the block to store.
// Every path that replaces blocks must call it, or a handle keeps resolving to a removed lodestone.
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
	// A lodestone rebuilt in the same spot gets a new handle, as in game, so UntrackPosition drops the old one.
	if existing := t.byPosition[key]; existing != 0 {
		handle = existing
	}
	if handle == 0 {
		t.next++
		handle = t.next
	} else if handle > t.next {
		// A handle loaded from a saved block raises the counter, so no later lodestone is issued it again.
		t.next = handle
	}
	if entry, ok := t.byHandle[handle]; ok {
		delete(t.byPosition, [4]int{entry.dim, entry.pos[0], entry.pos[1], entry.pos[2]})
	}
	t.byPosition[key] = handle
	t.byHandle[handle] = trackedPosition{pos: pos, dim: dim}
	delete(t.destroyed, handle)
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
	if handle != 0 {
		if t.destroyed == nil {
			t.destroyed = map[int32]struct{}{}
		}
		t.destroyed[handle] = struct{}{}
	}
	t.mu.Unlock()
	if handle == 0 {
		return
	}
	w.broadcastPositionTracking(pos, PositionTrackingDestroyAction{Handle: handle})
}

// BroadcastPositionTracking sends the target of a newly issued handle to every viewer of the world.
func (w *World) BroadcastPositionTracking(pos cube.Pos, handle int32) {
	dim, _ := DimensionID(w.Dimension())
	w.broadcastPositionTracking(pos, PositionTrackingUpdateAction{Handle: handle, Position: pos, Dimension: dim})
}

func (w *World) broadcastPositionTracking(pos cube.Pos, action BlockAction) {
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

// TrackedPosition returns the position a tracking handle resolves to and its status.
func (w *World) TrackedPosition(handle int32) (cube.Pos, int, PositionTrackingStatus) {
	t := w.set.tracker()
	t.mu.Lock()
	defer t.mu.Unlock()
	if entry, ok := t.byHandle[handle]; ok {
		return entry.pos, entry.dim, PositionTracked
	}
	if _, ok := t.destroyed[handle]; ok {
		return cube.Pos{}, 0, PositionDestroyed
	}
	return cube.Pos{}, 0, PositionNotFound
}
