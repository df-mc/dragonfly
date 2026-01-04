package world

import (
	"maps"
	"math"
	"sync"

	"github.com/go-gl/mathgl/mgl64"
)

// LoadStrategy determines which chunks should be loaded and when they should be unloaded.
type LoadStrategy interface {
	// Load returns the chunks that should be loaded, ordered by priority (closest first).
	Load(pos ChunkPos) []ChunkPos
	// Unload returns whether a chunk at the given position should be unloaded.
	Unload(chunk ChunkPos, pos ChunkPos) bool
}

// LoadRadius is a LoadStrategy that loads chunks in a circular radius around the current position.
type LoadRadius struct {
	Radius int
}

// Load returns chunks within the radius, ordered from center outwards.
func (s LoadRadius) Load(pos ChunkPos) []ChunkPos {
	queue := map[int32][]ChunkPos{}
	r := int32(s.Radius)

	for x := -r; x <= r; x++ {
		for z := -r; z <= r; z++ {
			distance := math.Sqrt(float64(x*x) + float64(z*z))
			chunkDistance := int32(math.Round(distance))
			if chunkDistance > r {
				// The chunk was outside the chunk radius.
				continue
			}
			chunkPos := ChunkPos{x + pos[0], z + pos[1]}
			queue[chunkDistance] = append(queue[chunkDistance], chunkPos)
		}
	}

	result := make([]ChunkPos, 0)
	for i := range r {
		result = append(result, queue[i]...)
	}
	return result
}

// Unload returns true if the chunk is outside the Radius.
func (s LoadRadius) Unload(chunk ChunkPos, pos ChunkPos) bool {
	diffX, diffZ := chunk[0]-pos[0], chunk[1]-pos[1]
	dist := math.Sqrt(float64(diffX*diffX) + float64(diffZ*diffZ))
	return int(dist) > s.Radius
}

// LoadArea is a LoadStrategy that loads chunks in a rectangular area around the current position.
type LoadArea struct {
	Width int
	Depth int
}

// Load returns chunks within the rectangular area, ordered from center outwards.
func (s LoadArea) Load(pos ChunkPos) []ChunkPos {
	queue := map[int32][]ChunkPos{}
	halfW, halfD := int32(s.Width/2), int32(s.Depth/2)

	for x := -halfW; x <= halfW; x++ {
		for z := -halfD; z <= halfD; z++ {
			distance := math.Sqrt(float64(x*x) + float64(z*z))
			chunkDistance := int32(math.Round(distance))
			chunkPos := ChunkPos{x + pos[0], z + pos[1]}
			queue[chunkDistance] = append(queue[chunkDistance], chunkPos)
		}
	}

	maxDist := int32(math.Sqrt(float64(halfW*halfW) + float64(halfD*halfD)))
	result := make([]ChunkPos, 0)
	for i := range maxDist {
		result = append(result, queue[i]...)
	}
	return result
}

// Unload returns true if the chunk is outside the Width and Depth.
func (s LoadArea) Unload(chunk ChunkPos, pos ChunkPos) bool {
	halfW, halfD := int32(s.Width/2), int32(s.Depth/2)
	diffX, diffZ := chunk[0]-pos[0], chunk[1]-pos[1]
	if diffX < 0 {
		diffX = -diffX
	}
	if diffZ < 0 {
		diffZ = -diffZ
	}
	return diffX > halfW || diffZ > halfD
}

// LoadRegion is a LoadStrategy that loads chunks within a fixed rectangular area defined by Min and Max
// chunk positions. It does not follow the current position and never unloads chunks.
type LoadRegion struct {
	Min, Max ChunkPos
}

// Load returns all chunks within the rectangular region.
func (s LoadRegion) Load(pos ChunkPos) []ChunkPos {
	result := make([]ChunkPos, 0)
	for x := s.Min[0]; x <= s.Max[0]; x++ {
		for z := s.Min[1]; z <= s.Max[1]; z++ {
			result = append(result, ChunkPos{x, z})
		}
	}
	return result
}

// Unload always returns false as LoadRegion never unloads chunks.
func (s LoadRegion) Unload(chunk ChunkPos, pos ChunkPos) bool {
	return false
}

// LoadManual is a LoadStrategy that allows manual control over which chunks to load and unload.
type LoadManual struct {
	mu     sync.RWMutex
	chunks map[ChunkPos]struct{}
}

// NewLoadManual returns a new LoadManual.
func NewLoadManual() *LoadManual {
	return &LoadManual{chunks: make(map[ChunkPos]struct{})}
}

// Add adds a chunk position to be loaded.
func (s *LoadManual) Add(pos ChunkPos) {
	s.mu.Lock()
	s.chunks[pos] = struct{}{}
	s.mu.Unlock()
}

// Remove removes a chunk position, marking it for unload.
func (s *LoadManual) Remove(pos ChunkPos) {
	s.mu.Lock()
	delete(s.chunks, pos)
	s.mu.Unlock()
}

// Load returns all manually added chunk positions.
func (s *LoadManual) Load(pos ChunkPos) []ChunkPos {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ChunkPos, 0, len(s.chunks))
	for chunk := range s.chunks {
		result = append(result, chunk)
	}
	return result
}

// Unload returns true if the chunk is not in the manual list.
func (s *LoadManual) Unload(chunk ChunkPos, pos ChunkPos) bool {
	s.mu.RLock()
	_, ok := s.chunks[chunk]
	s.mu.RUnlock()
	return !ok
}

// Loader implements the loading of the world. A loader can typically be moved around the world to load
// different parts of the world. An example usage is the player, which uses a loader to load chunks around it
// so that it can view them.
type Loader struct {
	strategy LoadStrategy
	w        *World
	viewer   Viewer

	mu        sync.RWMutex
	pos       ChunkPos
	loadQueue []ChunkPos
	loaded    map[ChunkPos]*Column

	closed bool
}

// NewLoader creates a new loader using the LoadStrategy passed. The strategy determines which chunks should
// be loaded and when they should be unloaded.
// The Viewer passed will handle the loading of chunks, including the viewing of entities that were loaded in
// those chunks.
func NewLoader(strategy LoadStrategy, world *World, v Viewer) *Loader {
	l := &Loader{strategy: strategy, loaded: make(map[ChunkPos]*Column), viewer: v}
	l.world(world)
	return l
}

// World returns the World that the Loader is in.
func (l *Loader) World() *World {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.w
}

// Strategy returns the LoadStrategy currently used by the Loader.
func (l *Loader) Strategy() LoadStrategy {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.strategy
}

// ChangeWorld changes the World of the Loader. The currently loaded chunks are reset and any future loading
// is done from the new World.
func (l *Loader) ChangeWorld(tx *Tx, new *World) {
	l.mu.Lock()
	defer l.mu.Unlock()

	loaded := maps.Clone(l.loaded)
	l.w.Exec(func(tx *Tx) {
		for pos := range loaded {
			tx.World().removeViewer(tx, pos, l)
		}
	})
	clear(l.loaded)
	l.w.viewerMu.Lock()
	delete(l.w.viewers, l)
	l.w.viewerMu.Unlock()

	l.world(new)
}

// ChangeStrategy changes the LoadStrategy of the Loader. Chunks that should no longer be loaded according to
// the new strategy will be unloaded, and new chunks will be queued for loading.
func (l *Loader) ChangeStrategy(tx *Tx, strategy LoadStrategy) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.strategy = strategy
	l.evictUnused(tx)
	l.populateLoadQueue()
}

// ChangeRadius changes the maximum chunk radius of the Loader if it uses LoadRadius. For other strategies,
// this method has no effect.
func (l *Loader) ChangeRadius(tx *Tx, new int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.strategy.(LoadRadius); ok {
		l.strategy = LoadRadius{Radius: new}
		l.evictUnused(tx)
		l.populateLoadQueue()
	}
}

// Refresh re-evaluates the current strategy, unloading chunks that should no longer be loaded and queuing
// new chunks for loading. This is useful after modifying a LoadManual.
func (l *Loader) Refresh(tx *Tx) {
	l.mu.Lock()
	defer l.mu.Unlock()

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

// Load loads n chunks around the centre of the chunk, starting with the middle and working outwards. For
// every chunk loaded, the Viewer passed through construction in New has its ViewChunk method called.
// Load does nothing for n <= 0.
func (l *Loader) Load(tx *Tx, n int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed || l.w == nil {
		return
	}
	for range n {
		if len(l.loadQueue) == 0 {
			break
		}

		pos := l.loadQueue[0]
		c := tx.w.chunk(pos)

		l.viewer.ViewChunk(pos, l.w.Dimension(), c.BlockEntities, c.Chunk)
		l.w.addViewer(tx, c, l)

		l.loaded[pos] = c

		// Shift the first element from the load queue off so that we can take a new one during the next
		// iteration.
		l.loadQueue = l.loadQueue[1:]
	}
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
		tx.World().removeViewer(tx, pos, l)
	}
	l.loaded = map[ChunkPos]*Column{}

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

// evictUnused gets rid of chunks in the loaded map which should no longer be loaded according to the strategy,
// and should therefore be removed.
func (l *Loader) evictUnused(tx *Tx) {
	for pos := range l.loaded {
		if l.strategy.Unload(pos, l.pos) {
			delete(l.loaded, pos)
			l.w.removeViewer(tx, pos, l)
		}
	}
}

// populateLoadQueue populates the load queue of the loader using the current strategy. Chunks that are
// already loaded are filtered out.
func (l *Loader) populateLoadQueue() {
	chunks := l.strategy.Load(l.pos)

	l.loadQueue = l.loadQueue[:0]
	for _, pos := range chunks {
		if _, ok := l.loaded[pos]; !ok {
			l.loadQueue = append(l.loadQueue, pos)
		}
	}
}
