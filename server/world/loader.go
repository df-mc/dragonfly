package world

import (
	"maps"
	"math"
	"sync"

	"github.com/go-gl/mathgl/mgl64"
)

// Loader implements the loading of the world. A loader can typically be moved around the world to load
// different parts of the world. An example usage is the player, which uses a loader to load chunks around it
// so that it can view them.
type Loader struct {
	r      int
	w      *World
	viewer Viewer

	mu        sync.RWMutex
	pos       ChunkPos
	loadQueue []ChunkPos
	loaded    map[ChunkPos]*Column
	pending   map[ChunkPos]struct{}

	closed bool
}

// NewLoader creates a new loader using the chunk radius passed. Chunks beyond this radius from the position
// of the loader will never be loaded.
// The Viewer passed will handle the loading of chunks, including the viewing of entities that were loaded in
// those chunks.
func NewLoader(chunkRadius int, world *World, v Viewer) *Loader {
	l := &Loader{r: chunkRadius, loaded: make(map[ChunkPos]*Column), pending: make(map[ChunkPos]struct{}), viewer: v}
	l.world(world)
	return l
}

// World returns the World that the Loader is in.
func (l *Loader) World() *World {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.w
}

// ChangeWorld changes the World of the Loader. The currently loaded chunks are reset and any future loading
// is done from the new World.
func (l *Loader) ChangeWorld(tx *Tx, new *World) {
	l.mu.Lock()
	defer l.mu.Unlock()

	loaded := maps.Clone(l.loaded)
	l.w.exec(func(tx *Tx) {
		for pos := range loaded {
			tx.World().removeViewer(tx, pos, l, l.viewer)
		}
	})
	clear(l.loaded)
	clear(l.pending)
	l.w.viewerMu.Lock()
	delete(l.w.viewers, l)
	l.w.viewerMu.Unlock()

	l.world(new)
}

// ChangeRadius changes the maximum chunk radius of the Loader.
func (l *Loader) ChangeRadius(tx *Tx, new int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.r = new
	l.evictUnused(tx)
	l.populateLoadQueue()
}

// Move moves the loader to the position passed. The position is translated to a chunk position to load
func (l *Loader) Move(tx *Tx, pos mgl64.Vec3) {
	l.mu.Lock()
	defer l.mu.Unlock()

	chunkPos := chunkPosFromVec3(pos)
	if chunkPos == l.pos {
		return
	}
	l.pos = chunkPos
	l.evictUnused(tx)
	l.populateLoadQueue()
}

// Load queues up to n chunks around the loader's centre, from the middle outwards, to be loaded in
// the background. The Viewer's ViewChunk is called for each chunk once ready, which may be after Load
// returns. Load does nothing for n <= 0.
func (l *Loader) Load(tx *Tx, n int) {
	for i := 0; i < n; i++ {
		l.mu.Lock()
		if l.closed || l.w == nil {
			l.mu.Unlock()
			return
		}
		if len(l.loadQueue) == 0 {
			l.mu.Unlock()
			break
		}
		pos := l.loadQueue[0]
		w := tx.World()
		l.pending[pos] = struct{}{}

		// Shift the first element from the load queue off so that we can take a new one during the next
		// iteration.
		l.loadQueue = l.loadQueue[1:]
		l.mu.Unlock()

		if !w.loadChunkAsync(tx, pos, func(tx2 *Tx, col *Column) {
			l.viewChunk(tx2, pos, col)
		}) {
			l.mu.Lock()
			delete(l.pending, pos)
			l.queueLoad(pos)
			l.mu.Unlock()
		}
	}
}

// viewChunk passes a loaded chunk to the Loader's Viewer. If the chunk failed
// to load, it is queued to be loaded again.
func (l *Loader) viewChunk(tx *Tx, pos ChunkPos, c *Column) {
	viewer, w, ok := l.recordChunk(tx, pos, c)
	if !ok {
		return
	}
	// The Viewer is called with l.mu released. Showing an entity to a Viewer may read a block outside the chunk being
	// viewed, which can complete a chunk request that another Load queued. Completing a request runs its callbacks
	// inline on this goroutine, and the callback a Loader registers is this method, so holding l.mu here would
	// deadlock the world's transaction goroutine against itself.
	viewer.ViewChunk(pos, w.Dimension(), c.BlockEntities, c.Chunk)
	w.addViewer(tx, c, l, viewer)

	if l.removeIfStale(w, pos) {
		// The Loader was closed or moved to a different World while the Viewer was being called, so it will never
		// remove itself from this chunk again. Undo the addition instead of leaking the Viewer into the old World.
		w.removeViewer(tx, pos, l, viewer)
	}
}

// recordChunk records a chunk as loaded by the Loader and returns the Viewer and World to pass it to. The last return
// value is false if the chunk should not be viewed at all.
//
// The chunk is recorded before the Viewer is called rather than after it, so that a re-entrant call for the same
// position returns at the loaded check below instead of blocking on a lock this goroutine already holds.
func (l *Loader) recordChunk(tx *Tx, pos ChunkPos, c *Column) (Viewer, *World, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed || l.viewer == nil || l.w == nil || l.w != tx.World() {
		return nil, nil, false
	}
	delete(l.pending, pos)
	if c == nil {
		l.queueLoad(pos)
		return nil, nil, false
	}
	if _, ok := l.loaded[pos]; ok {
		return nil, nil, false
	}
	if !l.withinLoadRadius(pos) {
		return nil, nil, false
	}
	l.loaded[pos] = c
	return l.viewer, l.w, true
}

// removeIfStale removes the chunk position passed from the chunks the Loader has loaded if the Loader was closed or
// moved to a different World than the one passed, and reports whether it did.
func (l *Loader) removeIfStale(w *World, pos ChunkPos) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.closed && l.w == w {
		return false
	}
	delete(l.loaded, pos)
	return true
}

// Chunk attempts to return a chunk at the given ChunkPos. If the chunk is not loaded, the second return value will
// be false.
func (l *Loader) Chunk(pos ChunkPos) (*Column, bool) {
	l.mu.RLock()
	c, ok := l.loaded[pos]
	l.mu.RUnlock()
	return c, ok
}

// Close closes the loader. It unloads all chunks currently loaded for the viewer, and hides all entities that
// are currently shown to it.
func (l *Loader) Close(tx *Tx) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for pos := range l.loaded {
		tx.World().removeViewer(tx, pos, l, l.viewer)
	}
	l.loaded = map[ChunkPos]*Column{}
	clear(l.pending)

	l.w.viewerMu.Lock()
	delete(l.w.viewers, l)
	l.w.viewerMu.Unlock()

	l.closed = true
	l.viewer = nil
}

// world sets the loader's world, adds them to the world's viewer list, then starts populating the load queue.
// This is only here to get rid of duplicated code, ChangeWorld should be used instead of this.
func (l *Loader) world(new *World) {
	l.w = new
	l.w.addWorldViewer(l)
	l.populateLoadQueue()
}

// evictUnused gets rid of chunks in the loaded map which are no longer within the chunk radius of the loader,
// and should therefore be removed.
func (l *Loader) evictUnused(tx *Tx) {
	for pos := range l.loaded {
		if !l.withinLoadRadius(pos) {
			delete(l.loaded, pos)
			l.w.removeViewer(tx, pos, l, l.viewer)
		}
	}
}

// withinLoadRadius checks if a chunk position is within the Loader's radius.
func (l *Loader) withinLoadRadius(pos ChunkPos) bool {
	return chunkDistance(pos, l.pos) <= int32(l.r)
}

// chunkDistance returns the rounded distance between two chunk positions.
func chunkDistance(a, b ChunkPos) int32 {
	diffX, diffZ := float64(a[0])-float64(b[0]), float64(a[1])-float64(b[1])
	return int32(math.Round(math.Sqrt(diffX*diffX + diffZ*diffZ)))
}

// queueLoad adds pos back to the load queue, unless it is already loaded,
// queued, or no longer within the radius of the Loader.
func (l *Loader) queueLoad(pos ChunkPos) {
	if l.closed || l.w == nil || !l.withinLoadRadius(pos) {
		return
	}
	if _, ok := l.loaded[pos]; ok {
		return
	}
	if _, ok := l.pending[pos]; ok {
		return
	}
	for _, queued := range l.loadQueue {
		if queued == pos {
			return
		}
	}
	l.loadQueue = append(l.loadQueue, pos)
}

// populateLoadQueue populates the load queue of the loader. This method is called once to create the order in
// which chunks around the position the loader is now in should be loaded. Chunks are ordered to be loaded
// from the middle outwards.
func (l *Loader) populateLoadQueue() {
	// We'll first load the chunk positions to load in a map indexed by the distance to the centre (basically,
	// what precedence it should have), and put them in the loadQueue in that order.
	queue := map[int32][]ChunkPos{}

	r := int32(l.r)
	for x := -r; x <= r; x++ {
		for z := -r; z <= r; z++ {
			pos := ChunkPos{x + l.pos[0], z + l.pos[1]}
			dist := chunkDistance(pos, l.pos)
			if dist > r {
				// The chunk was outside the chunk radius.
				continue
			}
			if _, ok := l.loaded[pos]; ok {
				// The chunk was already loaded, so we don't need to do anything.
				continue
			}
			if _, ok := l.pending[pos]; ok {
				// The chunk is already queued to be loaded.
				continue
			}
			queue[dist] = append(queue[dist], pos)
		}
	}

	l.loadQueue = l.loadQueue[:0]
	for i := int32(0); i <= r; i++ {
		l.loadQueue = append(l.loadQueue, queue[i]...)
	}
}
