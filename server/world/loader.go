package world

import (
	"github.com/go-gl/mathgl/mgl64"
	"math"
	"sync"
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
	loaded    map[ChunkPos]struct{}

	closed bool
}

// NewLoader creates a new loader using the chunk radius passed. Chunks beyond this radius from the position
// of the loader will never be loaded.
// The Viewer passed will handle the loading of chunks, including the viewing of entities that were loaded in
// those chunks.
func NewLoader(chunkRadius int, world *World, v Viewer) *Loader {
	l := &Loader{r: chunkRadius, loaded: make(map[ChunkPos]struct{}), viewer: v}
	l.world(world)
	return l
}

// World returns the World that the Loader is in.
func (l *Loader) World() *World {
	l.mu.RLock()
	w := l.w
	l.mu.RUnlock()
	return w
}

// ChangeWorld changes the World of the Loader. The currently loaded chunks are reset and any future loading
// is done from the new World.
func (l *Loader) ChangeWorld(new *World) {
	l.mu.Lock()
	l.reset()
	l.world(new)
	l.mu.Unlock()
}

// ChangeRadius changes the maximum chunk radius of the Loader.
func (l *Loader) ChangeRadius(new int) {
	l.mu.Lock()
	l.r = new

	l.evictUnused()
	l.populateLoadQueue()
	l.mu.Unlock()
}

// Move moves the loader to the position passed. The position is translated to a chunk position to load
func (l *Loader) Move(pos mgl64.Vec3) {
	l.mu.Lock()

	floorX, floorZ := math.Floor(pos[0]), math.Floor(pos[2])
	chunkPos := ChunkPos{int32(floorX) >> 4, int32(floorZ) >> 4}

	if chunkPos == l.pos {
		l.mu.Unlock()
		return
	}
	l.pos = chunkPos
	l.evictUnused()
	l.populateLoadQueue()

	l.mu.Unlock()
}

// Load loads n chunks around the centre of the chunk, starting with the middle and working outwards. For
// every chunk loaded, the function f is called.
// The function f must not hold the chunk beyond the function scope.
// An error is returned if one of the chunks could not be loaded.
func (l *Loader) Load(n int) error {
	if n == 0 {
		return nil
	}
	l.mu.Lock()
	if l.closed || l.w == nil {
		l.mu.Unlock()
		return nil
	}
	for i := 0; i < n; i++ {
		if len(l.loadQueue) == 0 {
			break
		}
		pos := l.loadQueue[0]
		c, err := l.w.chunk(pos)
		if err != nil {
			l.mu.Unlock()
			return err
		}
		l.viewer.ViewChunk(pos, c.Chunk, c.e)
		l.w.addViewer(c, l.viewer)

		l.loaded[pos] = struct{}{}

		// Shift the first element from the load queue off so that we can take a new one during the next
		// iteration.
		l.loadQueue = l.loadQueue[1:]
	}
	l.mu.Unlock()
	return nil
}

// Close closes the loader. It unloads all chunks currently loaded for the viewer, and hides all entities that
// are currently shown to it.
func (l *Loader) Close() error {
	l.mu.Lock()
	l.reset()
	l.closed = true
	l.viewer = nil
	l.mu.Unlock()
	return nil
}

// reset clears the Loader so that it may be used as if it was created again with NewLoader.
func (l *Loader) reset() {
	for pos := range l.loaded {
		l.w.removeViewer(pos, l.viewer)
	}
	l.loaded = map[ChunkPos]struct{}{}
	l.w.removeWorldViewer(l.viewer)
}

// world sets the loader's world, adds them to the world's viewer list, then starts populating the load queue.
// This is only here to get rid of duplicated code, ChangeWorld should be used instead of this.
func (l *Loader) world(new *World) {
	l.w = new
	l.w.addWorldViewer(l.viewer)
	l.populateLoadQueue()
}

// evictUnused gets rid of chunks in the loaded map which are no longer within the chunk radius of the loader,
// and should therefore be removed.
func (l *Loader) evictUnused() {
	for pos := range l.loaded {
		diffX, diffZ := pos[0]-l.pos[0], pos[1]-l.pos[1]
		dist := math.Sqrt(float64(diffX*diffX) + float64(diffZ*diffZ))
		if int(dist) > l.r {
			delete(l.loaded, pos)
			l.w.removeViewer(pos, l.viewer)
		}
	}
}

// populateLoadQueue populates the load queue of the loader. This method is called once to create the order in
// which chunks around the position the loader is now in should be loaded. Chunks are ordered to be loaded
// from the middle outwards.
func (l *Loader) populateLoadQueue() {
	l.loadQueue = nil
	// We'll first load the chunk positions to load in a map indexed by the distance to the center (basically,
	// what precedence it should have), and put them in the loadQueue in that order.
	toLoad := map[int32][]ChunkPos{}

	chunkX, chunkZ := l.pos[0], l.pos[1]
	r := int32(l.r)

	for x := -r; x <= r; x++ {
		for z := -r; z <= r; z++ {
			distance := math.Sqrt(float64(x*x) + float64(z*z))
			chunkDistance := int32(math.Round(distance))
			if chunkDistance > r {
				// The chunk was outside the chunk radius.
				continue
			}
			pos := ChunkPos{x + chunkX, z + chunkZ}
			if _, ok := l.loaded[pos]; ok {
				// The chunk was already loaded, so we don't need to do anything.
				continue
			}
			if m, ok := toLoad[chunkDistance]; ok {
				toLoad[chunkDistance] = append(m, pos)
				continue
			}
			toLoad[chunkDistance] = []ChunkPos{pos}
		}
	}
	for i := int32(0); i < r; i++ {
		l.loadQueue = append(l.loadQueue, toLoad[i]...)
	}
}
