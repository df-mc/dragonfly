package world

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/chunk"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"sync"
)

// Loader implements the loading of the world. A loader can typically be moved around the world to load
// different parts of the world. An example usage is the player, which uses a loader to load chunks around it
// so that it can view them.
type Loader struct {
	r int
	w *World

	mutex     sync.RWMutex
	pos       ChunkPos
	loadQueue []ChunkPos
	loaded    map[ChunkPos]*chunk.Chunk
}

// NewLoader creates a new loader using the chunk radius passed. Chunks beyond this radius from the position
// of the loader will never be loaded.
func NewLoader(chunkRadius int, world *World) *Loader {
	return &Loader{r: chunkRadius, w: world, loaded: make(map[ChunkPos]*chunk.Chunk)}
}

// Move moves the loader to the position passed. The position is translated to a chunk position to load
func (l *Loader) Move(pos mgl32.Vec3) {
	floorX, floorZ := math.Floor(float64(pos[0])), math.Floor(float64(pos[2]))
	l.mutex.Lock()
	l.pos = ChunkPos{int32(floorX) >> 4, int32(floorZ) >> 4}
	l.mutex.Unlock()
	l.populateLoadQueue()
}

// Load loads n chunks around the centre of the chunk, starting with the middle and working outwards. For
// every chunk loaded, the function f is called.
// The function f must not hold the chunk beyond the function scope.
// An error is returned if one of the chunks could not be loaded.
func (l *Loader) Load(n int, f func(pos ChunkPos, c *chunk.Chunk)) error {
	l.mutex.Lock()
	for i := 0; i < n; i++ {
		if len(l.loadQueue) == 0 {
			l.mutex.Unlock()
			return nil
		}
		c, err := l.w.chunk(l.loadQueue[0])
		if err != nil {
			l.mutex.Unlock()
			return err
		}
		f(l.loadQueue[0], c)
		c.Unlock()

		l.loaded[l.loadQueue[0]] = c

		// Shift the first element from the load queue off so that we can take a new one during the next
		// iteration.
		l.loadQueue = l.loadQueue[1:]
	}
	l.mutex.Unlock()
	return nil
}

// populateLoadQueue populates the load queue of the loader. This method is called once to create the order in
// which chunks around the position the loader is now in should be loaded. Chunks are ordered to be loaded
// from the middle outwards.
func (l *Loader) populateLoadQueue() {
	l.mutex.Lock()
	l.loadQueue = nil
	// We'll first load the chunk positions to load in a map indexed by the distance to the center (basically,
	// what precedence it should have), and put them in the loadQueue in that order.
	toLoad := map[int32][]ChunkPos{}

	chunkX, chunkZ := l.pos[0], l.pos[1]
	r := int32(l.r)

	for x := -r; x <= r; x++ {
		for z := -r; z <= r; z++ {
			pos := ChunkPos{x + chunkX, z + chunkZ}
			if _, ok := l.loaded[pos]; ok {
				// The chunk was already loaded, so we don't need to do anything.
				continue
			}
			distance := math.Sqrt(float64(x*x) + float64(z*z))
			chunkDistance := int32(math.Round(distance))
			if chunkDistance > int32(l.r) {
				// The chunk was outside of the chunk radius.
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
	l.mutex.Unlock()
}
