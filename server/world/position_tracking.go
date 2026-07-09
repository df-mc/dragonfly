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
	pos    cube.Pos
	dim    int
	active bool
}

// PositionTrackingEntry is a persistent entry in the position tracking database.
type PositionTrackingEntry struct {
	Handle    int32
	Position  cube.Pos
	Dimension int
	Active    bool
}

// PositionTrackingDestroyAction notifies viewers that a tracked block no
// longer exists, causing matching compasses to spin immediately.
type PositionTrackingDestroyAction struct {
	Handle int32
}

// BlockAction implements BlockAction.
func (PositionTrackingDestroyAction) BlockAction() {}

// PositionTrackingUpdateAction provides the target of a tracking handle to a
// viewer immediately after a compass is linked.
type PositionTrackingUpdateAction struct {
	Handle    int32
	Position  cube.Pos
	Dimension int
}

// BlockAction implements BlockAction.
func (PositionTrackingUpdateAction) BlockAction() {}

type positionTracker struct {
	mu         sync.Mutex
	next       int32
	byHandle   map[int32]trackedPosition
	byPosition map[[4]int]int32
}

// PositionTrackingEntries returns a snapshot of the position tracking database.
func (s *Settings) PositionTrackingEntries() []PositionTrackingEntry {
	t := &s.positionTracker
	t.mu.Lock()
	defer t.mu.Unlock()
	entries := make([]PositionTrackingEntry, 0, len(t.byHandle))
	for handle, entry := range t.byHandle {
		entries = append(entries, PositionTrackingEntry{Handle: handle, Position: entry.pos, Dimension: entry.dim, Active: entry.active})
	}
	return entries
}

// LoadPositionTrackingEntries replaces the position tracking database with entries.
func (s *Settings) LoadPositionTrackingEntries(entries []PositionTrackingEntry) {
	t := &s.positionTracker
	t.mu.Lock()
	defer t.mu.Unlock()
	t.next = 0
	t.byHandle = map[int32]trackedPosition{}
	t.byPosition = map[[4]int]int32{}
	for _, entry := range entries {
		if entry.Handle == 0 {
			continue
		}
		t.byHandle[entry.Handle] = trackedPosition{pos: entry.Position, dim: entry.Dimension, active: entry.Active}
		t.byPosition[[4]int{entry.Dimension, entry.Position[0], entry.Position[1], entry.Position[2]}] = entry.Handle
		if entry.Handle > t.next {
			t.next = entry.Handle
		}
	}
}

// TrackPosition activates a tracking handle for pos. Existing handles at the
// same position are reused so a lodestone replaced before a stored compass is
// read remains linked, matching Bedrock behaviour.
func (w *World) TrackPosition(pos cube.Pos, handle int32) int32 {
	dim, _ := DimensionID(w.Dimension())
	t := &w.set.positionTracker
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.byHandle == nil {
		t.byHandle = map[int32]trackedPosition{}
		t.byPosition = map[[4]int]int32{}
	}
	key := [4]int{dim, pos[0], pos[1], pos[2]}
	if existing := t.byPosition[key]; existing != 0 {
		handle = existing
	}
	if handle == 0 {
		for handle == 0 || t.byHandle[handle].active {
			t.next++
			handle = t.next
		}
	}
	if entry, ok := t.byHandle[handle]; ok {
		delete(t.byPosition, [4]int{entry.dim, entry.pos[0], entry.pos[1], entry.pos[2]})
	}
	t.byPosition[key] = handle
	t.byHandle[handle] = trackedPosition{pos: pos, dim: dim, active: true}
	return handle
}

// UntrackPosition marks the tracking handle at pos as unavailable. Its position
// association is retained so replacing the lodestone can reactivate it.
func (w *World) UntrackPosition(pos cube.Pos) {
	dim, _ := DimensionID(w.Dimension())
	t := &w.set.positionTracker
	t.mu.Lock()
	handle := t.byPosition[[4]int{dim, pos[0], pos[1], pos[2]}]
	if handle != 0 {
		entry := t.byHandle[handle]
		entry.active = false
		t.byHandle[handle] = entry
	}
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

// TrackedPosition looks up an active position tracking handle.
func (w *World) TrackedPosition(handle int32) (cube.Pos, int, bool) {
	t := &w.set.positionTracker
	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.byHandle[handle]
	if !ok {
		return cube.Pos{}, 0, false
	}
	if !entry.active {
		return cube.Pos{}, 0, false
	}
	return entry.pos, entry.dim, true
}
