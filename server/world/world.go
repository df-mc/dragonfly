package world

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/internal"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
	"go.uber.org/atomic"
	"math/rand"
	"sync"
	"time"
)

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
// World generally provides a synchronised state: All entities, blocks and players usually operate in this
// world, so World ensures that all its methods will always be safe for simultaneous calls.
// A nil *World is safe to use but not functional.
type World struct {
	log internal.Logger

	mu   sync.Mutex
	set  Settings
	prov Provider

	rdonly atomic.Bool

	lastPos   ChunkPos
	lastChunk *chunkData

	closing chan struct{}
	running sync.WaitGroup

	handlerMu sync.RWMutex
	handler   Handler

	genMu sync.RWMutex
	gen   Generator

	chunkMu sync.Mutex
	// chunks holds a cache of chunks currently loaded. These chunks are cleared from this map after some time
	// of not being used.
	chunks map[ChunkPos]*chunkData

	entityMu sync.RWMutex
	// entities holds a map of entities currently loaded and the last ChunkPos that the Entity was in.
	// These are tracked so that a call to RemoveEntity can find the correct entity.
	entities map[Entity]ChunkPos

	r         *rand.Rand
	simDistSq int32

	randomTickSpeed atomic.Uint32

	updateMu sync.Mutex
	// blockUpdates is a map of tick time values indexed by the block position at which an update is
	// scheduled. If the current tick exceeds the tick value passed, the block update will be performed
	// and the entry will be removed from the map.
	blockUpdates             map[cube.Pos]int64
	updatePositions          []cube.Pos
	neighbourUpdatePositions []neighbourUpdate
	neighbourUpdatesSync     []neighbourUpdate

	toTick              []toTick
	blockEntitiesToTick []blockEntityToTick
	positionCache       []ChunkPos
	entitiesToTick      []TickerEntity

	viewersMu sync.Mutex
	viewers   map[Viewer]struct{}
}

// New creates a new initialised world. The world may be used right away, but it will not be saved or loaded
// from files until it has been given a different provider than the default. (NoIOProvider)
// By default, the name of the world will be 'World'.
func New(log internal.Logger, simulationDistance int) *World {
	w := &World{
		r:               rand.New(rand.NewSource(time.Now().Unix())),
		blockUpdates:    map[cube.Pos]int64{},
		entities:        map[Entity]ChunkPos{},
		viewers:         map[Viewer]struct{}{},
		prov:            NoIOProvider{},
		gen:             NopGenerator{},
		handler:         NopHandler{},
		simDistSq:       int32(simulationDistance * simulationDistance),
		randomTickSpeed: *atomic.NewUint32(3),
		log:             log,
		set:             defaultSettings(),
		closing:         make(chan struct{}),
	}

	w.initChunkCache()
	go w.startTicking()
	go w.chunkCacheJanitor()
	return w
}

// Name returns the display name of the world. Generally, this name is displayed at the top of the player list
// in the pause screen in-game.
// If a provider is set, the name will be updated according to the name that it provides.
func (w *World) Name() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.set.Name
}

// Block reads a block from the position passed. If a chunk is not yet loaded at that position, the chunk is
// loaded, or generated if it could not be found in the world save, and the block returned. Chunks will be
// loaded synchronously.
func (w *World) Block(pos cube.Pos) Block {
	if w == nil || pos.OutOfBounds() {
		// Fast way out.
		return air()
	}
	chunkPos := ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)}
	c, err := w.chunk(chunkPos)
	if err != nil {
		w.log.Errorf("error getting block: %v", err)
		return air()
	}
	rid := c.RuntimeID(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)

	b, _ := BlockByRuntimeID(rid)
	if nbtBlocks[rid] {
		// The block was also a block entity, so we look it up in the block entity map.
		if nbtB, ok := c.e[pos]; ok {
			c.Unlock()
			return nbtB
		}
	}
	c.Unlock()

	return b
}

// blockInChunk reads a block from the world at the position passed. The block is assumed to be in the chunk
// passed, which is also assumed to be locked already or otherwise not yet accessible.
func (w *World) blockInChunk(c *chunkData, pos cube.Pos) (Block, error) {
	if pos.OutOfBounds() {
		// Fast way out.
		return air(), nil
	}
	rid := c.RuntimeID(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
	b, _ := BlockByRuntimeID(rid)

	if nbtBlocks[rid] {
		// The block was also a block entity, so we look it up in the block entity map.
		b, ok := c.e[pos]
		if ok {
			return b, nil
		}
	}
	return b, nil
}

// runtimeID gets the block runtime ID at a specific position in the world.
//lint:ignore U1000 Function is used using compiler directives.
//noinspection GoUnusedFunction
func runtimeID(w *World, pos cube.Pos) uint32 {
	if w == nil || pos.OutOfBounds() {
		// Fast way out.
		return airRID
	}
	c, err := w.chunk(ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)})
	if err != nil {
		return airRID
	}
	rid := c.RuntimeID(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
	c.Unlock()

	return rid
}

// HighestLightBlocker gets the Y value of the highest fully light blocking block at the x and z values
// passed in the world.
func (w *World) HighestLightBlocker(x, z int) int16 {
	if w == nil {
		return cube.MinY
	}
	c, err := w.chunk(ChunkPos{int32(x >> 4), int32(z >> 4)})
	if err != nil {
		return cube.MinY
	}
	v := c.HighestLightBlocker(uint8(x), uint8(z))
	c.Unlock()
	return v
}

// HighestBlock looks up the highest non-air block in the world at a specific x and z in the world. The y
// value of the highest block is returned, or 0 if no blocks were present in the column.
func (w *World) HighestBlock(x, z int) int {
	if w == nil {
		return cube.MinY
	}
	c, err := w.chunk(ChunkPos{int32(x >> 4), int32(z >> 4)})
	if err != nil {
		return cube.MinY
	}
	v := c.HighestBlock(uint8(x), uint8(z))
	c.Unlock()
	return int(v)
}

// highestObstructingBlock returns the highest block in the world at a given x and z that has at least a solid top or
// bottom face.
func (w *World) highestObstructingBlock(x, z int) int {
	if w == nil {
		return 0
	}
	yHigh := w.HighestBlock(x, z)
	for y := yHigh; y >= cube.MinY; y-- {
		pos := cube.Pos{x, y, z}
		m := w.Block(pos).Model()
		if m.FaceSolid(pos, cube.FaceUp, w) || m.FaceSolid(pos, cube.FaceDown, w) {
			return y
		}
	}
	return cube.MinY
}

// SetBlock writes a block to the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
// SetBlock panics if the block passed has not yet been registered using RegisterBlock().
// Nil may be passed as the block to set the block to air.
// SetBlock should be avoided in situations where performance is critical when needing to set a lot of blocks
// to the world. BuildStructure may be used instead.
func (w *World) SetBlock(pos cube.Pos, b Block) {
	if w == nil || pos.OutOfBounds() {
		// Fast way out.
		return
	}

	x, z := int32(pos[0]>>4), int32(pos[2]>>4)
	c, err := w.chunk(ChunkPos{x, z})
	if err != nil {
		return
	}

	rid, ok := BlockRuntimeID(b)
	if !ok {
		c.Unlock()
		w.log.Errorf("runtime ID of block %+v not found", b)
		return
	}
	c.SetRuntimeID(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0, rid)

	if nbtBlocks[rid] {
		c.e[pos] = b
	} else {
		delete(c.e, pos)
	}

	var viewers []Viewer
	if len(c.v) > 0 {
		viewers = make([]Viewer, len(c.v))
		copy(viewers, c.v)
	}
	c.Unlock()

	for _, viewer := range viewers {
		viewer.ViewBlockUpdate(pos, b, 0)
	}
}

// breakParticle has its value set in the block_internal package.
var breakParticle func(b Block) Particle

// BreakBlock breaks a block at the position passed. Unlike when setting the block at that position to air,
// BreakBlock will also show particles and update blocks around the position.
func (w *World) BreakBlock(pos cube.Pos) {
	if w == nil {
		return
	}
	old := w.Block(pos)
	w.SetBlock(pos, nil)
	w.AddParticle(pos.Vec3Centre(), breakParticle(old))
	if liq, ok := w.Liquid(pos); ok {
		// Move the liquid down a layer.
		w.SetLiquid(pos, liq)
	} else {
		w.doBlockUpdatesAround(pos)
	}
}

// BreakBlockWithoutParticles breaks a block at the position passed. Unlike when setting the block at that position to air,
// BreakBlockWithoutParticles will also update blocks around the position.
func (w *World) BreakBlockWithoutParticles(pos cube.Pos) {
	if w == nil {
		return
	}
	w.SetBlock(pos, nil)
	if liq, ok := w.Liquid(pos); ok {
		// Move the liquid down a layer.
		w.SetLiquid(pos, liq)
	} else {
		w.doBlockUpdatesAround(pos)
	}
}

// PlaceBlock places a block at the position passed. Unlike when using SetBlock, PlaceBlock also schedules
// block updates around the position.
// If the block can displace liquids at the position placed, it will do so, and liquid source blocks will be
// put into the same block as the one passed.
func (w *World) PlaceBlock(pos cube.Pos, b Block) {
	if w == nil {
		return
	}
	var liquid Liquid
	if displacer, ok := b.(LiquidDisplacer); ok {
		liq, ok := w.Liquid(pos)
		if ok && displacer.CanDisplace(liq) && liq.LiquidDepth() == 8 {
			liquid = liq
		}
	}
	w.SetBlock(pos, b)
	w.SetLiquid(pos, liquid)
}

// BuildStructure builds a Structure passed at a specific position in the world. Unlike SetBlock, it takes a
// Structure implementation, which provides blocks to be placed at a specific location.
// BuildStructure is specifically tinkered to be able to process a large batch of chunks simultaneously and
// will do so within much less time than separate SetBlock calls would.
// The method operates on a per-chunk basis, setting all blocks within a single chunk part of the structure
// before moving on to the next chunk.
func (w *World) BuildStructure(pos cube.Pos, s Structure) {
	if w == nil {
		return
	}
	dim := s.Dimensions()
	width, height, length := dim[0], dim[1], dim[2]
	maxX, maxY, maxZ := pos[0]+width, pos[1]+height, pos[2]+length

	for chunkX := pos[0] >> 4; chunkX <= maxX>>4; chunkX++ {
		for chunkZ := pos[2] >> 4; chunkZ <= maxZ>>4; chunkZ++ {
			// We approach this on a per-chunk basis, so that we can keep only one chunk in memory at a time
			// while not needing to acquire a new chunk lock for every block. This also allows us not to send
			// block updates, but instead send a single chunk update once.
			chunkPos := ChunkPos{int32(chunkX), int32(chunkZ)}
			c, err := w.chunk(chunkPos)
			if err != nil {
				w.log.Errorf("error loading chunk for structure: %v", err)
				continue
			}
			f := func(x, y, z int) Block {
				actualX, actualY, actualZ := pos[0]+x, pos[1]+y, pos[2]+z
				if actualX>>4 == chunkX && actualZ>>4 == chunkZ {
					b, _ := w.blockInChunk(c, cube.Pos{actualX, actualY, actualZ})
					return b
				}
				return w.Block(cube.Pos{actualX, actualY, actualZ})
			}
			baseX, baseZ := chunkX<<4, chunkZ<<4
			subs := c.Sub()
			for i, sub := range subs {
				baseY := (i + (cube.MinY >> 4)) << 4
				if sub == nil {
					c.SetRuntimeID(0, int16(baseY), 0, 0, airRID)
					sub = subs[i]
				}

				if baseY>>4 < pos[1]>>4 {
					continue
				} else if baseY >= maxY {
					break
				}

				for localY := 0; localY < 16; localY++ {
					yOffset := baseY + localY
					if yOffset > cube.MaxY || yOffset >= maxY {
						// We've hit the height limit for blocks.
						break
					} else if yOffset < cube.MinY || yOffset < pos[1] {
						// We've got a block below the minimum, but other blocks might still reach above
						// it, so don't break but continue.
						continue
					}
					for localX := 0; localX < 16; localX++ {
						xOffset := baseX + localX
						if xOffset < pos[0] || xOffset >= maxX {
							continue
						}
						for localZ := 0; localZ < 16; localZ++ {
							zOffset := baseZ + localZ
							if zOffset < pos[2] || zOffset >= maxZ {
								continue
							}
							b, liq := s.At(xOffset-pos[0], yOffset-pos[1], zOffset-pos[2], f)
							if b != nil {
								rid, ok := BlockRuntimeID(b)
								if !ok {
									w.log.Errorf("error setting block of structure: runtime ID of block state %+v not found", b)
									continue
								}
								sub.SetRuntimeID(uint8(xOffset), uint8(yOffset), uint8(zOffset), 0, rid)

								if nbtBlocks[rid] {
									c.e[pos] = b
								} else {
									delete(c.e, pos)
								}
							}
							if liq != nil {
								rid, ok := BlockRuntimeID(liq)
								if !ok {
									w.log.Errorf("runtime ID of block state %+v not found", liq)
									continue
								}
								sub.SetRuntimeID(uint8(xOffset), uint8(yOffset), uint8(zOffset), 1, rid)
							} else {
								if len(sub.Layers()) < 2 {
									continue
								}
								sub.SetRuntimeID(uint8(xOffset), uint8(yOffset), uint8(zOffset), 1, airRID)
							}
						}
					}
				}
			}
			// After setting all blocks of the structure within a single chunk, we show the new chunk to all
			// viewers once, and unlock it.
			for _, viewer := range c.v {
				viewer.ViewChunk(chunkPos, c.Chunk, c.e)
			}
			c.Unlock()
		}
	}
}

// Liquid attempts to return any liquid block at the position passed. This liquid may be in the foreground or
// in any other layer.
// If found, the liquid is returned. If not, the bool returned is false and the liquid is nil.
func (w *World) Liquid(pos cube.Pos) (Liquid, bool) {
	if w == nil || pos.OutOfBounds() {
		// Fast way out.
		return nil, false
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos))
	if err != nil {
		w.log.Errorf("failed getting liquid: error getting chunk at position %v: %v", chunkPosFromBlockPos(pos), err)
		return nil, false
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])

	id := c.RuntimeID(x, y, z, 0)
	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.log.Errorf("failed getting liquid: cannot get block by runtime ID %v", id)
		c.Unlock()
		return nil, false
	}
	if liq, ok := b.(Liquid); ok {
		c.Unlock()
		return liq, true
	}

	id = c.RuntimeID(x, y, z, 1)
	b, ok = BlockByRuntimeID(id)
	c.Unlock()
	if !ok {
		w.log.Errorf("failed getting liquid: cannot get block by runtime ID %v", id)
		return nil, false
	}
	if liq, ok := b.(Liquid); ok {
		return liq, true
	}
	return nil, false
}

// SetLiquid sets the liquid at a specific position in the world. Unlike SetBlock, SetLiquid will not
// overwrite any existing blocks. It will instead be in the same position as a block currently there, unless
// there already is a liquid at that position, in which case it will be overwritten.
// If nil is passed for the liquid, any liquid currently present will be removed.
func (w *World) SetLiquid(pos cube.Pos, b Liquid) {
	if w == nil || pos.OutOfBounds() {
		// Fast way out.
		return
	}
	chunkPos := chunkPosFromBlockPos(pos)
	c, err := w.chunk(chunkPos)
	if err != nil {
		w.log.Errorf("failed setting liquid: error getting chunk at position %v: %v", chunkPosFromBlockPos(pos), err)
		return
	}
	if b == nil {
		w.removeLiquids(c, pos)
		c.Unlock()
		w.doBlockUpdatesAround(pos)
		return
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	if !replaceable(w, c, pos, b) {
		current, err := w.blockInChunk(c, pos)
		if err != nil {
			c.Unlock()
			w.log.Errorf("failed setting liquid: error getting block at position %v: %v", chunkPosFromBlockPos(pos), err)
			return
		}
		if displacer, ok := current.(LiquidDisplacer); !ok || !displacer.CanDisplace(b) {
			c.Unlock()
			return
		}
	}
	runtimeID, ok := BlockRuntimeID(b)
	if !ok {
		c.Unlock()
		w.log.Errorf("failed setting liquid: runtime ID of block state %+v not found", b)
		return
	}
	if w.removeLiquids(c, pos) {
		c.SetRuntimeID(x, y, z, 0, runtimeID)
		for _, v := range c.v {
			v.ViewBlockUpdate(pos, b, 0)
		}
	} else {
		c.SetRuntimeID(x, y, z, 1, runtimeID)
		for _, v := range c.v {
			v.ViewBlockUpdate(pos, b, 1)
		}
	}
	c.Unlock()

	w.doBlockUpdatesAround(pos)
}

// removeLiquids removes any liquid blocks that may be present at a specific block position in the chunk
// passed.
// The bool returned specifies if no blocks were left on the foreground layer.
func (w *World) removeLiquids(c *chunkData, pos cube.Pos) bool {
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])

	noneLeft := false
	if noLeft, changed := w.removeLiquidOnLayer(c.Chunk, x, y, z, 0); noLeft {
		if changed {
			for _, v := range c.v {
				v.ViewBlockUpdate(pos, air(), 0)
			}
		}
		noneLeft = true
	}
	if _, changed := w.removeLiquidOnLayer(c.Chunk, x, y, z, 1); changed {
		for _, v := range c.v {
			v.ViewBlockUpdate(pos, air(), 1)
		}
	}
	return noneLeft
}

// removeLiquidOnLayer removes a liquid block from a specific layer in the chunk passed, returning true if
// successful.
func (w *World) removeLiquidOnLayer(c *chunk.Chunk, x uint8, y int16, z, layer uint8) (bool, bool) {
	id := c.RuntimeID(x, y, z, layer)

	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.log.Errorf("failed removing liquids: cannot get block by runtime ID %v", id)
		return false, false
	}
	if _, ok := b.(Liquid); ok {
		c.SetRuntimeID(x, y, z, layer, airRID)
		return true, true
	}
	return id == airRID, false
}

// additionalLiquid checks if the block at a position has additional liquid on another layer and returns the
// liquid if so.
func (w *World) additionalLiquid(pos cube.Pos) (Liquid, bool) {
	if pos.OutOfBounds() {
		// Fast way out.
		return nil, false
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos))
	if err != nil {
		w.log.Errorf("failed getting liquid: error getting chunk at position %v: %v", chunkPosFromBlockPos(pos), err)
		return nil, false
	}
	id := c.RuntimeID(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 1)
	c.Unlock()
	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.log.Errorf("failed getting liquid: cannot get block by runtime ID %v", id)
		return nil, false
	}
	liq, ok := b.(Liquid)
	return liq, ok
}

// Light returns the light level at the position passed. This is the highest of the sky and block light.
// The light value returned is a value in the range 0-15, where 0 means there is no light present, whereas
// 15 means the block is fully lit.
func (w *World) Light(pos cube.Pos) uint8 {
	if w == nil || pos[1] < cube.MinY {
		// Fast way out.
		return 0
	}
	if pos[1] > cube.MaxY {
		// Above the rest of the world, so full sky light.
		return 15
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos))
	if err != nil {
		return 0
	}
	l := c.Light(uint8(pos[0]), int16(pos[1]), uint8(pos[2]))
	c.Unlock()

	return l
}

// SkyLight returns the sky light level at the position passed. This light level is not influenced by blocks
// that emit light, such as torches or glowstone. The light value, similarly to Light, is a value in the
// range 0-15, where 0 means no light is present.
func (w *World) SkyLight(pos cube.Pos) uint8 {
	if w == nil || pos[1] < cube.MinY {
		// Fast way out.
		return 0
	}
	if pos[1] > cube.MaxY {
		// Above the rest of the world, so full sky light.
		return 15
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos))
	if err != nil {
		return 0
	}
	l := c.SkyLight(uint8(pos[0]), int16(pos[1]), uint8(pos[2]))
	c.Unlock()

	return l
}

// Time returns the current time of the world. The time is incremented every 1/20th of a second, unless
// World.StopTime() is called.
func (w *World) Time() int {
	if w == nil {
		return 0
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return int(w.set.Time)
}

// SetTime sets the new time of the world. SetTime will always work, regardless of whether the time is stopped
// or not.
func (w *World) SetTime(new int) {
	if w == nil {
		return
	}
	w.mu.Lock()
	w.set.Time = int64(new)
	w.mu.Unlock()
	for _, viewer := range w.allViewers() {
		viewer.ViewTime(new)
	}
}

// StopTime stops the time in the world. When called, the time will no longer cycle and the world will remain
// at the time when StopTime is called. The time may be restarted by calling World.StartTime().
// StopTime will not do anything if the time is already stopped.
func (w *World) StopTime() {
	w.enableTimeCycle(false)
}

// StartTime restarts the time in the world. When called, the time will start cycling again and the day/night
// cycle will continue. The time may be stopped again by calling World.StopTime().
// StartTime will not do anything if the time is already started.
func (w *World) StartTime() {
	w.enableTimeCycle(true)
}

// enableTimeCycle enables or disables the time cycling of the World.
func (w *World) enableTimeCycle(v bool) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.set.TimeCycle = v
}

// StopWeatherCycle disables weather of the World.
func (w *World) StopWeatherCycle() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.set.WeatherCycle = false
}

// StartWeatherCycle enables weather of the World.
func (w *World) StartWeatherCycle() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.set.WeatherCycle = true
}

// RainingAt returns a bool that indicates whether it is raining at a position in the world.
// TODO: Take into account biomes when deciding if it is raining at a position when we implement biomes.
func (w *World) RainingAt(pos cube.Pos) bool {
	if w == nil {
		return false
	}
	w.mu.Lock()
	a := w.set.Raining
	w.mu.Unlock()
	return a && w.highestObstructingBlock(pos[0], pos[2]) < pos[1]
}

// ThunderingAt returns a bool indicating whether it is currently thundering or not. True is returned only if it is both
// raining and thundering at the same time and if the position passed is exposed to rain.
// TODO: Take into account biomes when deciding if it is thundering at a position when we implement biomes.
func (w *World) ThunderingAt(pos cube.Pos) bool {
	if w == nil {
		return false
	}
	w.mu.Lock()
	a := w.set.Thundering && w.set.Raining
	w.mu.Unlock()
	return a && w.highestObstructingBlock(pos[0], pos[2]) < pos[1]
}

// AddParticle spawns a particle at a given position in the world. Viewers that are viewing the chunk will be
// shown the particle.
func (w *World) AddParticle(pos mgl64.Vec3, p Particle) {
	if w == nil {
		return
	}
	p.Spawn(w, pos)
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewParticle(pos, p)
	}
}

// PlaySound plays a sound at a specific position in the world. Viewers of that position will be able to hear
// the sound if they're close enough.
func (w *World) PlaySound(pos mgl64.Vec3, s Sound) {
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewSound(pos, s)
	}
}

var (
	worldsMu sync.RWMutex
	// entityWorlds holds a list of all entities added to a world. It may be used to lookup the world that an
	// entity is currently in.
	entityWorlds = map[Entity]*World{}
)

// AddEntity adds an entity to the world at the position that the entity has. The entity will be visible to
// all viewers of the world that have the chunk of the entity loaded.
// If the chunk that the entity is in is not yet loaded, it will first be loaded.
// If the entity passed to AddEntity is currently in a world, it is first removed from that world.
func (w *World) AddEntity(e Entity) {
	if w == nil {
		return
	}
	if e.World() != nil {
		e.World().RemoveEntity(e)
	}
	worldsMu.Lock()
	entityWorlds[e] = w
	worldsMu.Unlock()

	chunkPos := chunkPosFromVec3(e.Position())
	w.entityMu.Lock()
	w.entities[e] = chunkPos
	w.entityMu.Unlock()

	c, err := w.chunk(chunkPos)
	if err != nil {
		w.log.Errorf("error loading chunk to add entity: %v", err)
		return
	}
	c.entities = append(c.entities, e)

	var viewers []Viewer
	if len(c.v) > 0 {
		viewers = make([]Viewer, len(c.v))
		copy(viewers, c.v)
	}
	c.Unlock()

	for _, viewer := range viewers {
		// We show the entity to all viewers currently in the chunk that the entity is spawned in.
		showEntity(e, viewer)
	}
}

// RemoveEntity removes an entity from the world that is currently present in it. Any viewers of the entity
// will no longer be able to see it.
// RemoveEntity operates assuming the position of the entity is the same as where it is currently in the
// world. If it can not find it there, it will loop through all entities and try to find it.
// RemoveEntity assumes the entity is currently loaded and in a loaded chunk. If not, the function will not do
// anything.
func (w *World) RemoveEntity(e Entity) {
	if w == nil {
		return
	}
	w.entityMu.Lock()
	chunkPos, found := w.entities[e]
	if !found {
		w.entityMu.Unlock()
		// The entity currently isn't in this world.
		return
	}
	w.entityMu.Unlock()

	worldsMu.Lock()
	delete(entityWorlds, e)
	worldsMu.Unlock()

	c, ok := w.chunkFromCache(chunkPos)
	if !ok {
		// The chunk wasn't loaded, so we can't remove any entity from the chunk.
		return
	}
	c.Lock()
	n := make([]Entity, 0, len(c.entities))
	for _, entity := range c.entities {
		if entity != e {
			n = append(n, entity)
			continue
		}
	}
	c.entities = n

	var viewers []Viewer
	if len(c.v) > 0 {
		viewers = make([]Viewer, len(c.v))
		copy(viewers, c.v)
	}
	c.Unlock()

	w.entityMu.Lock()
	delete(w.entities, e)
	w.entityMu.Unlock()

	for _, viewer := range viewers {
		viewer.HideEntity(e)
	}
}

// EntitiesNearby returns a list of entities that are intersecting the bounding box given.
func (w *World) EntitiesNearby(aabb physics.AABB, ignored func(Entity) bool) []Entity {
	if w == nil {
		return nil
	}
	// Make an estimate of 16 entities on average.
	m := make([]Entity, 0, 16)

	minPos, maxPos := chunkPosFromVec3(aabb.Min()), chunkPosFromVec3(aabb.Max())

	for x := minPos[0]; x <= maxPos[0]; x++ {
		for z := minPos[1]; z <= maxPos[1]; z++ {
			c, ok := w.chunkFromCache(ChunkPos{x, z})
			if !ok {
				// The chunk wasn't loaded, so there are no entities here.
				continue
			}
			c.Lock()
			for _, entity := range c.entities {
				if ignored != nil && ignored(entity) {
					continue
				}
				if aabb.IntersectsWith(entity.AABB().Translate(entity.Position())) {
					// The entity position intersected the AABB, so we add it to the slice to return.
					m = append(m, entity)
				}
			}
			c.Unlock()
		}
	}
	return m
}

// EntitiesWithin does a lookup through the entities in the chunks touched by the AABB passed, returning all
// those which are contained within the AABB when it comes to their position.
func (w *World) EntitiesWithin(aabb physics.AABB, ignored func(Entity) bool) []Entity {
	if w == nil {
		return nil
	}
	// Make an estimate of 16 entities on average.
	m := make([]Entity, 0, 16)

	minPos, maxPos := chunkPosFromVec3(aabb.Min()), chunkPosFromVec3(aabb.Max())

	for x := minPos[0]; x <= maxPos[0]; x++ {
		for z := minPos[1]; z <= maxPos[1]; z++ {
			c, ok := w.chunkFromCache(ChunkPos{x, z})
			if !ok {
				// The chunk wasn't loaded, so there are no entities here.
				continue
			}
			c.Lock()
			for _, entity := range c.entities {
				if ignored != nil && ignored(entity) {
					continue
				}
				if aabb.Vec3Within(entity.Position()) {
					// The entity position was within the AABB, so we add it to the slice to return.
					m = append(m, entity)
				}
			}
			c.Unlock()
		}
	}
	return m
}

// Entities returns a list of all entities currently added to the World.
func (w *World) Entities() []Entity {
	if w == nil {
		return nil
	}
	w.entityMu.RLock()
	m := make([]Entity, 0, len(w.entities))
	for e := range w.entities {
		m = append(m, e)
	}
	w.entityMu.RUnlock()
	return m
}

// OfEntity attempts to return a world that an entity is currently in. If the entity was not currently added
// to a world, the world returned is nil and the bool returned is false.
func OfEntity(e Entity) (*World, bool) {
	worldsMu.RLock()
	w, ok := entityWorlds[e]
	worldsMu.RUnlock()
	return w, ok
}

// Spawn returns the spawn of the world. Every new player will by default spawn on this position in the world
// when joining.
func (w *World) Spawn() cube.Pos {
	if w == nil {
		return cube.Pos{}
	}
	w.mu.Lock()
	s := w.set.Spawn
	w.mu.Unlock()
	if s[1] > cube.MaxY {
		s[1] = w.highestObstructingBlock(s[0], s[2]) + 1
	}
	return s
}

// SetSpawn sets the spawn of the world to a different position. The player will be spawned in the center of
// this position when newly joining.
func (w *World) SetSpawn(pos cube.Pos) {
	if w == nil {
		return
	}
	w.mu.Lock()
	w.set.Spawn = pos
	w.mu.Unlock()
	for _, viewer := range w.allViewers() {
		viewer.ViewWorldSpawn(pos)
	}
}

// DefaultGameMode returns the default game mode of the world. When players join, they are given this game
// mode.
// The default game mode may be changed using SetDefaultGameMode().
func (w *World) DefaultGameMode() GameMode {
	if w == nil {
		return GameModeSurvival
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.set.DefaultGameMode
}

// SetDefaultGameMode changes the default game mode of the world. When players join, they are then given that
// game mode.
func (w *World) SetDefaultGameMode(mode GameMode) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.set.DefaultGameMode = mode
}

// Difficulty returns the difficulty of the world. Properties of mobs in the world and the player's hunger
// will depend on this difficulty.
func (w *World) Difficulty() Difficulty {
	if w == nil {
		return DifficultyNormal{}
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.set.Difficulty
}

// SetDifficulty changes the difficulty of a world.
func (w *World) SetDifficulty(d Difficulty) {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.set.Difficulty = d
}

// SetRandomTickSpeed sets the random tick speed of blocks. By default, each sub chunk has 3 blocks randomly
// ticked per sub chunk, so the default value is 3. Setting this value to 0 will stop random ticking
// altogether, while setting it higher results in faster ticking.
func (w *World) SetRandomTickSpeed(v int) {
	if w == nil {
		return
	}
	w.randomTickSpeed.Store(uint32(v))
}

// ScheduleBlockUpdate schedules a block update at the position passed after a specific delay. If the block at
// that position does not handle block updates, nothing will happen.
func (w *World) ScheduleBlockUpdate(pos cube.Pos, delay time.Duration) {
	if w == nil || pos.OutOfBounds() {
		return
	}
	w.updateMu.Lock()
	if _, exists := w.blockUpdates[pos]; exists {
		w.updateMu.Unlock()
		return
	}
	w.mu.Lock()
	t := w.set.CurrentTick
	w.mu.Unlock()

	w.blockUpdates[pos] = t + delay.Nanoseconds()/int64(time.Second/20)
	w.updateMu.Unlock()
}

// doBlockUpdatesAround schedules block updates directly around and on the position passed.
func (w *World) doBlockUpdatesAround(pos cube.Pos) {
	if w == nil || pos.OutOfBounds() {
		return
	}

	changed := pos

	w.updateMu.Lock()
	w.updateNeighbour(pos, changed)
	pos.Neighbours(func(pos cube.Pos) {
		w.updateNeighbour(pos, changed)
	})
	w.updateMu.Unlock()
}

// neighbourUpdate represents a position that needs to be updated because of a neighbour that changed.
type neighbourUpdate struct {
	pos, neighbour cube.Pos
}

// updateNeighbour ticks the position passed as a result of the neighbour passed being updated.
func (w *World) updateNeighbour(pos, changedNeighbour cube.Pos) {
	w.neighbourUpdatePositions = append(w.neighbourUpdatePositions, neighbourUpdate{pos: pos, neighbour: changedNeighbour})
}

// Provider changes the provider of the world to the provider passed. If nil is passed, the NoIOProvider
// will be set, which does not read or write any data.
func (w *World) Provider(p Provider) {
	if w == nil {
		return
	}
	if p == nil {
		p = NoIOProvider{}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.set = p.Settings()
	w.prov = p

	w.initChunkCache()
}

// ReadOnly makes the world read only. Chunks will no longer be saved to disk, just like entities and data
// in the level.dat.
func (w *World) ReadOnly() {
	if w == nil {
		return
	}
	w.rdonly.Store(true)
}

// Generator changes the generator of the world to the one passed. If nil is passed, the generator is set to
// the default, NopGenerator.
func (w *World) Generator(g Generator) {
	if w == nil {
		return
	}
	w.genMu.Lock()
	defer w.genMu.Unlock()

	if g == nil {
		g = NopGenerator{}
	}
	w.gen = g
}

// Handle changes the current Handler of the world. As a result, events called by the world will call
// handlers of the Handler passed.
// Handle sets the world's Handler to NopHandler if nil is passed.
func (w *World) Handle(h Handler) {
	if w == nil {
		return
	}
	w.handlerMu.Lock()
	defer w.handlerMu.Unlock()

	if h == nil {
		h = NopHandler{}
	}
	w.handler = h
}

// Viewers returns a list of all viewers viewing the position passed. A viewer will be assumed to be watching
// if the position is within one of the chunks that the viewer is watching.
func (w *World) Viewers(pos mgl64.Vec3) (viewers []Viewer) {
	if w == nil {
		return nil
	}
	c, ok := w.chunkFromCache(chunkPosFromVec3(pos))
	if !ok {
		return nil
	}
	c.Lock()
	if len(c.v) > 0 {
		viewers = make([]Viewer, len(c.v))
		copy(viewers, c.v)
	}
	c.Unlock()
	return
}

// Close closes the world and saves all chunks currently loaded.
func (w *World) Close() error {
	if w == nil {
		return nil
	}
	close(w.closing)
	w.running.Wait()

	w.log.Debugf("Saving chunks in memory to disk...")

	w.chunkMu.Lock()
	w.lastChunk = nil
	chunksToSave := make(map[ChunkPos]*chunkData, len(w.chunks))
	for pos, c := range w.chunks {
		// We delete all chunks from the cache and save them to the provider.
		delete(w.chunks, pos)
		chunksToSave[pos] = c
	}
	w.chunkMu.Unlock()

	for pos, c := range chunksToSave {
		w.saveChunk(pos, c)
	}

	if !w.rdonly.Load() {
		w.log.Debugf("Updating level.dat values...")
		w.provider().SaveSettings(w.set)
	}

	w.log.Debugf("Closing provider...")
	if err := w.provider().Close(); err != nil {
		w.log.Errorf("error closing world provider: %v", err)
	}
	w.Handle(NopHandler{})
	return nil
}

// startTicking starts ticking the world, updating all entities, blocks and other features such as the time of
// the world, as required.
func (w *World) startTicking() {
	ticker := time.NewTicker(time.Second / 20)
	defer ticker.Stop()

	w.running.Add(1)
	for {
		select {
		case <-ticker.C:
			w.tick()
		case <-w.closing:
			// World is being closed: Stop ticking and get rid of a task.
			w.running.Done()
			return
		}
	}
}

// tick ticks the world and updates the time, blocks and entities that require updates.
func (w *World) tick() {
	viewers := w.allViewers()
	if len(viewers) == 0 {
		return
	}

	w.mu.Lock()
	tick := w.set.CurrentTick
	w.set.CurrentTick++

	if w.set.TimeCycle {
		w.set.Time++
	}
	t := int(w.set.Time)

	if w.set.WeatherCycle {
		w.set.RainTime--
		if w.set.RainTime <= 0 {
			// Wiki: The rain counter counts down to zero, and each time it reaches zero, the rain is toggled on or off.
			// When the rain is turned on, the counter is reset to a value between 12,000-23,999 ticks (0.5-1 game days)
			// and when the rain is turned off it is reset to a value of 12,000-179,999 ticks (0.5-7.5 game days).
			if w.set.Raining {
				w.setRaining(false, time.Second*(time.Duration(w.r.Intn(8400)+600)))
			} else {
				w.setRaining(true, time.Second*time.Duration(w.r.Intn(600)+600))
			}
		}
		w.set.ThunderTime--
		if w.set.ThunderTime <= 0 {
			// Wiki: the thunder counter toggles thunder on/off when it reaches zero, but clear weather overrides the
			// "on" state. When thunder is turned on, the thunder counter is reset to 3,600-15,999 ticks (3-13 minutes),
			// and when thunder is turned off the counter rests to 12,000-179,999 ticks (0.5-7.5 days).
			if w.set.Thundering {
				w.setThunder(false, time.Second*(time.Duration(w.r.Intn(8400)+600)))
			} else {
				w.setThunder(true, time.Second*time.Duration(w.r.Intn(620)+180))
			}
		}
	}
	thunder := w.set.Thundering && w.set.Raining
	w.mu.Unlock()

	if thunder {
		w.chunkMu.Lock()
		positions := make([]ChunkPos, 0, len(w.chunks)/100000)
		for pos := range w.chunks {
			// Wiki: For each loaded chunk, every tick there is a 1⁄100,000 chance of an attempted lightning strike
			// during a thunderstorm
			if w.r.Intn(100000) == 0 {
				positions = append(positions, pos)
			}
		}
		w.chunkMu.Unlock()

		for _, pos := range positions {
			w.strikeLightning(pos)
		}
	}

	if tick%20 == 0 {
		for _, viewer := range viewers {
			viewer.ViewTime(t)
		}
	}

	w.tickEntities(tick)
	w.tickRandomBlocks(viewers, tick)
	w.tickScheduledBlocks(tick)
}

// strikeLightning attempts to strike lightning in the world at a specific ChunkPos. The final position is influenced by
// living entities that might be near the lightning strike. If there is no rain at the final position selected, the
// lightning strike will fail.
func (w *World) strikeLightning(c ChunkPos) {
	v := int32(w.r.Uint32())
	x, z := float64(c[0]<<4+(v&0xf)), float64(c[1]<<4+((v>>8)&0xf))

	vec := mgl64.Vec3{x, float64(w.HighestBlock(int(x), int(z)) + 1), z}
	ent := w.EntitiesWithin(physics.NewAABB(vec, vec.Add(mgl64.Vec3{0, 255})).GrowVec3(mgl64.Vec3{3, 3, 3}), nil)

	list := make([]mgl64.Vec3, 0, len(ent)/3)
	for _, e := range ent {
		if h, ok := e.(interface{ Health() float64 }); ok && h.Health() > 0 {
			// Any (living) entity that is positioned higher than the highest block at its position is eligible to be
			// struck by lightning. We first save all entity positions where this is the case.
			pos := cube.PosFromVec3(e.Position())
			if w.HighestBlock(pos[0], pos[1]) < pos[2] {
				list = append(list, e.Position())
			}
		}
	}
	// We then select one of the positions of entities higher than the highest block and adjust the position of the
	// lightning to it, so that the entity is struck directly.
	if len(list) > 0 {
		vec = list[w.r.Intn(len(list))]
	}

	pos := cube.PosFromVec3(vec)
	if len(w.Block(pos).Model().AABB(pos, w)) != 0 {
		// If lightning is about to strike inside of a block that is not fully transparent. In this case, move the
		// lightning up by one block so that it strikes above the block.
		vec = vec.Add(mgl64.Vec3{0, 1})
	}
	if !w.ThunderingAt(pos) {
		// No thunder at this position, meaning we were either under an obstructing block or in a biome where it does
		// not rain.
		return
	}

	e, _ := EntityByName("minecraft:lightning_bolt")
	w.AddEntity(e.(interface {
		New(mgl64.Vec3) Entity
	}).New(vec))
}

// tickScheduledBlocks executes scheduled block ticks in chunks that are still loaded at the time of
// execution.
func (w *World) tickScheduledBlocks(tick int64) {
	w.updateMu.Lock()
	for pos, scheduledTick := range w.blockUpdates {
		if scheduledTick <= tick {
			w.updatePositions = append(w.updatePositions, pos)
			delete(w.blockUpdates, pos)
		}
	}
	w.neighbourUpdatesSync = append(w.neighbourUpdatesSync, w.neighbourUpdatePositions...)
	w.neighbourUpdatePositions = w.neighbourUpdatePositions[:0]
	w.updateMu.Unlock()

	for _, pos := range w.updatePositions {
		if ticker, ok := w.Block(pos).(ScheduledTicker); ok {
			ticker.ScheduledTick(pos, w, w.r)
		}
		if liquid, ok := w.additionalLiquid(pos); ok {
			if ticker, ok := liquid.(ScheduledTicker); ok {
				ticker.ScheduledTick(pos, w, w.r)
			}
		}
	}
	for _, update := range w.neighbourUpdatesSync {
		pos, changedNeighbour := update.pos, update.neighbour
		if ticker, ok := w.Block(pos).(NeighbourUpdateTicker); ok {
			ticker.NeighbourUpdateTick(pos, changedNeighbour, w)
		}
		if liquid, ok := w.additionalLiquid(pos); ok {
			if ticker, ok := liquid.(NeighbourUpdateTicker); ok {
				ticker.NeighbourUpdateTick(pos, changedNeighbour, w)
			}
		}
	}

	w.updatePositions = w.updatePositions[:0]
	w.neighbourUpdatesSync = w.neighbourUpdatesSync[:0]
}

// toTick is a struct used to keep track of blocks that need to be ticked upon a random tick.
type toTick struct {
	b   RandomTicker
	pos cube.Pos
}

// blockEntityToTick is a struct used to keep track of block entities that need to be ticked upon a normal
// world tick.
type blockEntityToTick struct {
	b   TickerBlock
	pos cube.Pos
}

// tickRandomBlocks executes random block ticks in each sub chunk in the world that has at least one viewer
// registered from the viewers passed.
func (w *World) tickRandomBlocks(viewers []Viewer, tick int64) {
	if w.simDistSq == 0 {
		// NOP if the simulation distance is 0.
		return
	}
	tickSpeed := w.randomTickSpeed.Load()

	for _, viewer := range viewers {
		pos := viewer.Position()
		w.positionCache = append(w.positionCache, ChunkPos{
			// Technically we could obtain the wrong chunk position here due to truncating, but this
			// inaccuracy doesn't matter and it allows us to cut a corner.
			int32(pos[0]) >> 4,
			int32(pos[2]) >> 4,
		})
	}

	var g randUint4

	w.chunkMu.Lock()
	for pos, c := range w.chunks {
		withinSimDist := false
		for _, chunkPos := range w.positionCache {
			xDiff, zDiff := chunkPos[0]-pos[0], chunkPos[1]-pos[1]
			if (xDiff*xDiff)+(zDiff*zDiff) <= w.simDistSq {
				// The chunk was within the simulation distance of at least one viewer, so we can proceed to
				// ticking the block.
				withinSimDist = true
				break
			}
		}
		if !withinSimDist {
			// No viewers in this chunk that are within the simulation distance, so proceed to the next.
			continue
		}
		c.Lock()
		for pos, b := range c.e {
			if ticker, ok := b.(TickerBlock); ok {
				w.blockEntitiesToTick = append(w.blockEntitiesToTick, blockEntityToTick{
					b:   ticker,
					pos: pos,
				})
			}
		}

		subChunks := c.Sub()
		cx, cz := int(pos[0]<<4), int(pos[1]<<4)

		// We generate a random block in every chunk
		for j := uint32(0); j < tickSpeed; j++ {
			generateNew := true
			var x, y, z uint8
			for i, sub := range subChunks {
				if sub == nil {
					// No sub chunk present, so skip it right away.
					continue
				}
				layers := sub.Layers()
				if len(layers) == 0 {
					// No layers present, so skip it right away.
					continue
				}
				layer := layers[0]
				if p := layer.Palette(); p.Len() == 1 && p.RuntimeID(0) == airRID {
					// Empty layer present, so skip it right away.
					continue
				}
				if generateNew {
					x, y, z = g.uint4(w.r), g.uint4(w.r), g.uint4(w.r)
				}
				// Generally we would want to make sure the block has its block entities, but provided blocks
				// with block entities are generally ticked already, we are safe to assume that blocks
				// implementing the RandomTicker don't rely on additional block entity data.
				rid := layer.RuntimeID(x, y, z)
				if rid == airRID {
					// The block was air, take the fast route out.
					continue
				}

				if randomTickBlocks[rid] {
					subY := (i + (cube.MinY >> 4)) << 4
					w.toTick = append(w.toTick, toTick{b: blocks[rid].(RandomTicker), pos: cube.Pos{cx + int(x), subY + int(y), cz + int(z)}})
					generateNew = true
					continue
				}
				// No block at this position that was a RandomTicker. In this case, we can actually just move one sub
				// chunk up and try again, without generating any new positions. This one wasn't used, after all.
				generateNew = false
			}
		}
		c.Unlock()
	}
	w.chunkMu.Unlock()

	for _, a := range w.toTick {
		a.b.RandomTick(a.pos, w, w.r)
	}
	for _, b := range w.blockEntitiesToTick {
		b.b.Tick(tick, b.pos, w)
	}
	w.toTick = w.toTick[:0]
	w.blockEntitiesToTick = w.blockEntitiesToTick[:0]
	w.positionCache = w.positionCache[:0]
}

// randUint4 is a structure used to generate random uint4s.
type randUint4 struct {
	x uint64
	n uint8
}

// uint4 returns a random uint4.
func (g *randUint4) uint4(r *rand.Rand) uint8 {
	if g.n == 0 {
		g.x = r.Uint64()
		g.n = 16
	}
	val := g.x & 0b1111

	g.x >>= 4
	g.n--
	return uint8(val)
}

// tickEntities ticks all entities in the world, making sure they are still located in the correct chunks and
// updating where necessary.
func (w *World) tickEntities(tick int64) {
	type entityToMove struct {
		e             Entity
		after         *chunkData
		viewersBefore []Viewer
	}
	var entitiesToMove []entityToMove

	w.entityMu.Lock()
	w.chunkMu.Lock()
	for e, lastPos := range w.entities {
		chunkPos := chunkPosFromVec3(e.Position())

		c, ok := w.chunks[chunkPos]
		if !ok {
			continue
		}

		c.Lock()
		v := len(c.v)
		c.Unlock()

		if v > 0 {
			if ticker, ok := e.(TickerEntity); ok {
				w.entitiesToTick = append(w.entitiesToTick, ticker)
			}
		}

		if lastPos != chunkPos {
			// The entity was stored using an outdated chunk position. We update it and make sure it is ready
			// for viewers to view it.
			w.entities[e] = chunkPos

			old := w.chunks[lastPos]
			old.Lock()
			chunkEntities := make([]Entity, 0, len(old.entities)-1)
			for _, entity := range old.entities {
				if entity == e {
					continue
				}
				chunkEntities = append(chunkEntities, entity)
			}
			old.entities = chunkEntities

			var viewers []Viewer
			if len(old.v) > 0 {
				viewers = make([]Viewer, len(old.v))
				copy(viewers, old.v)
			}
			old.Unlock()

			entitiesToMove = append(entitiesToMove, entityToMove{e: e, viewersBefore: viewers, after: c})
		}
	}
	w.chunkMu.Unlock()
	w.entityMu.Unlock()

	for _, move := range entitiesToMove {
		move.after.Lock()
		move.after.entities = append(move.after.entities, move.e)
		viewersAfter := move.after.v
		move.after.Unlock()

		for _, viewer := range move.viewersBefore {
			if !w.hasViewer(viewer, viewersAfter) {
				// First we hide the entity from all viewers that were previously viewing it, but no
				// longer are.
				viewer.HideEntity(move.e)
			}
		}
		for _, viewer := range viewersAfter {
			if !w.hasViewer(viewer, move.viewersBefore) {
				// Then we show the entity to all viewers that are now viewing the entity in the new
				// chunk.
				showEntity(move.e, viewer)
			}
		}
	}
	for _, ticker := range w.entitiesToTick {
		if _, ok := OfEntity(ticker.(Entity)); !ok {
			continue
		}
		// We gather entities to tick and tick them later, so that the lock on the entity mutex is no longer
		// active.
		ticker.Tick(tick)
	}
	w.entitiesToTick = w.entitiesToTick[:0]
}

// StartRaining makes it rain in the current world where the time.Duration passed will determine how long it will rain.
func (w *World) StartRaining(dur time.Duration) {
	w.mu.Lock()
	w.setRaining(true, dur)
	w.mu.Unlock()
}

// StopRaining makes it stop raining in the current world.
func (w *World) StopRaining() {
	w.mu.Lock()
	if w.set.Raining {
		w.setRaining(false, time.Second*(time.Duration(w.r.Intn(8400)+600)))
		if w.set.Thundering {
			// Also reset thunder if it was previously thundering.
			w.setThunder(false, time.Second*(time.Duration(w.r.Intn(8400)+600)))
		}
	}
	w.mu.Unlock()
}

// setRaining toggles raining depending on the raining argument.
// This does not lock the world mutex as opposed to StartRaining and StopRaining.
func (w *World) setRaining(raining bool, x time.Duration) {
	w.set.Raining = raining
	w.set.RainTime = int64(x.Seconds() * 20)
	for _, v := range w.allViewers() {
		v.ViewWeather(raining, w.set.Raining && w.set.Thundering)
	}
}

// StartThundering makes it thunder in the current world where the time.Duration passed will determine how long it will
// thunder. StartThundering will also make it rain.
func (w *World) StartThundering(dur time.Duration) {
	w.mu.Lock()
	w.setThunder(true, dur)
	w.setRaining(true, dur)
	w.mu.Unlock()
}

// StopThundering makes it stop thundering in the current world.
func (w *World) StopThundering() {
	w.mu.Lock()
	if w.set.Thundering && w.set.Raining {
		w.setThunder(false, time.Second*(time.Duration(w.r.Intn(8400)+600)))
	}
	w.mu.Unlock()
}

// setThunder toggles thundering depending on the thundering argument.
// This does not lock the world mutex as opposed to StartThundering and StopThundering.
func (w *World) setThunder(thundering bool, x time.Duration) {
	w.set.Thundering = thundering
	w.set.ThunderTime = int64(x.Seconds() * 20)
	for _, v := range w.allViewers() {
		// Thunderstorms only happen if it is already raining. Clear weather overrides thunder, so only make it thunder
		// if it's both raining and thundering.
		v.ViewWeather(w.set.Raining, w.set.Raining && thundering)
	}
}

// allViewers returns a list of all viewers of the world, regardless of where in the world they are viewing.
func (w *World) allViewers() (v []Viewer) {
	w.viewersMu.Lock()
	if len(w.viewers) > 0 {
		v = make([]Viewer, 0, len(w.viewers))
		for viewer := range w.viewers {
			v = append(v, viewer)
		}
	}
	w.viewersMu.Unlock()
	return
}

// addWorldViewer adds a viewer to the world. Should only be used while the viewer isn't viewing any chunks.
func (w *World) addWorldViewer(viewer Viewer) {
	w.viewersMu.Lock()
	w.viewers[viewer] = struct{}{}
	w.viewersMu.Unlock()
	viewer.ViewTime(w.Time())
	w.mu.Lock()
	raining, thundering := w.set.Raining, w.set.Raining && w.set.Thundering
	w.mu.Unlock()
	viewer.ViewWeather(raining, thundering)
	viewer.ViewWorldSpawn(w.Spawn())
}

// removeWorldViewer removes a viewer from the world. Should only be used while the viewer isn't viewing any chunks.
func (w *World) removeWorldViewer(viewer Viewer) {
	w.viewersMu.Lock()
	delete(w.viewers, viewer)
	w.viewersMu.Unlock()
}

// addViewer adds a viewer to the world at a given position. Any events that happen in the chunk at that
// position, such as block changes, entity changes etc., will be sent to the viewer.
func (w *World) addViewer(c *chunkData, viewer Viewer) {
	if w == nil {
		return
	}
	c.v = append(c.v, viewer)

	var entities []Entity
	if len(c.entities) > 0 {
		entities = make([]Entity, len(c.entities))
		copy(entities, c.entities)
	}
	c.Unlock()

	for _, entity := range entities {
		showEntity(entity, viewer)
	}
}

// removeViewer removes a viewer from the world at a given position. All entities will be hidden from the
// viewer and no more calls will be made when events in the chunk happen.
func (w *World) removeViewer(pos ChunkPos, viewer Viewer) {
	if w == nil {
		return
	}
	c, ok := w.chunkFromCache(pos)
	if !ok {
		return
	}
	c.Lock()
	n := make([]Viewer, 0, len(c.v))
	for _, v := range c.v {
		if v != viewer {
			// Add all viewers but the one to remove to the new viewers slice.
			n = append(n, v)
		}
	}
	c.v = n

	var entities []Entity
	if len(c.entities) > 0 {
		entities = make([]Entity, len(c.entities))
		copy(entities, c.entities)
	}
	c.Unlock()

	// After removing the viewer from the chunk, we also need to hide all entities from the viewer.
	for _, entity := range entities {
		viewer.HideEntity(entity)
	}
}

// hasViewer checks if a chunk at a particular chunk position has the viewer passed. If so, true is returned.
func (w *World) hasViewer(viewer Viewer, viewers []Viewer) bool {
	if w == nil {
		return false
	}
	for _, v := range viewers {
		if v == viewer {
			return true
		}
	}
	return false
}

// provider returns the provider of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) provider() Provider {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.prov
}

// Handler returns the Handler of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) Handler() Handler {
	if w == nil {
		return NopHandler{}
	}
	w.handlerMu.RLock()
	handler := w.handler
	w.handlerMu.RUnlock()
	return handler
}

// generator returns the generator of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) generator() Generator {
	w.genMu.RLock()
	generator := w.gen
	w.genMu.RUnlock()
	return generator
}

// chunkFromCache attempts to fetch a chunk at the chunk position passed from the cache. If not found, the
// chunk returned is nil and false is returned.
func (w *World) chunkFromCache(pos ChunkPos) (*chunkData, bool) {
	w.chunkMu.Lock()
	c, ok := w.chunks[pos]
	w.chunkMu.Unlock()
	return c, ok
}

// showEntity shows an entity to a viewer of the world. It makes sure everything of the entity, including the
// items held, is shown.
func showEntity(e Entity, viewer Viewer) {
	viewer.ViewEntity(e)
	viewer.ViewEntityState(e)
	viewer.ViewEntityItems(e)
	viewer.ViewEntityArmour(e)
}

// chunk reads a chunk from the position passed. If a chunk at that position is not yet loaded, the chunk is
// loaded from the provider, or generated if it did not yet exist. Both of these actions are done
// synchronously.
// An error is returned if the chunk could not be loaded successfully.
// chunk locks the chunk returned, meaning that any call to chunk made at the same time has to wait until the
// user calls Chunk.Unlock() on the chunk returned.
func (w *World) chunk(pos ChunkPos) (*chunkData, error) {
	w.chunkMu.Lock()
	if pos == w.lastPos && w.lastChunk != nil {
		c := w.lastChunk
		w.chunkMu.Unlock()
		c.Lock()
		return c, nil
	}
	c, ok := w.chunks[pos]
	if !ok {
		var err error
		c, err = w.loadChunk(pos)
		if err != nil {
			return nil, err
		}

		c.Unlock()
		w.chunkMu.Lock()
		w.calculateLight(c.Chunk, pos)
	}
	w.lastChunk, w.lastPos = c, pos
	w.chunkMu.Unlock()

	c.Lock()
	return c, nil
}

// setChunk sets the chunk.Chunk passed at a specific ChunkPos without replacing any entities at that
// position.
//lint:ignore U1000 This method is explicitly present to be used using compiler directives.
func (w *World) setChunk(pos ChunkPos, c *chunk.Chunk) {
	if w == nil {
		return
	}
	w.chunkMu.Lock()
	defer w.chunkMu.Unlock()

	data, ok := w.chunks[pos]
	if ok {
		data.Chunk = c
	} else {
		data = newChunkData(c)
		w.chunks[pos] = data
	}
	blockNBT := make([]map[string]interface{}, 0, len(c.BlockNBT()))
	for pos, e := range c.BlockNBT() {
		e["x"], e["y"], e["z"] = int32(pos[0]), int32(pos[1]), int32(pos[2])
		blockNBT = append(blockNBT, e)
	}
	w.loadIntoBlocks(data, blockNBT)
}

// loadChunk attempts to load a chunk from the provider, or generates a chunk if one doesn't currently exist.
func (w *World) loadChunk(pos ChunkPos) (*chunkData, error) {
	c, found, err := w.provider().LoadChunk(pos)
	if err != nil {
		return nil, fmt.Errorf("error loading chunk %v: %w", pos, err)
	}

	if !found {
		// The provider doesn't have a chunk saved at this position, so we generate a new one.
		c = chunk.New(airRID)
		data := newChunkData(c)
		w.chunks[pos] = data
		data.Lock()
		w.chunkMu.Unlock()

		w.generator().GenerateChunk(pos, c)
		for _, sub := range c.Sub() {
			if sub == nil {
				continue
			}
			// Creating new sub chunks will create a fully lit sub chunk, but we don't want that here as
			// light updates aren't happening (yet).
			sub.ClearLight()
		}
		return data, nil
	}
	data := newChunkData(c)
	w.chunks[pos] = data
	data.Lock()
	w.chunkMu.Unlock()

	ent, err := w.provider().LoadEntities(pos)
	if err != nil {
		return nil, fmt.Errorf("error loading entities of chunk %v: %w", pos, err)
	}
	data.entities = make([]Entity, 0, len(ent))

	// Iterate through the entities twice and make sure they're added to all relevant maps. Note that this iteration
	// happens twice to avoid having to lock both worldsMu and entityMu. This is intentional, to avoid deadlocks.
	worldsMu.Lock()
	for _, e := range ent {
		data.entities = append(data.entities, e)
		entityWorlds[e] = w
	}
	worldsMu.Unlock()

	w.entityMu.Lock()
	for _, e := range ent {
		w.entities[e] = pos
	}
	w.entityMu.Unlock()

	blockEntities, err := w.provider().LoadBlockNBT(pos)
	if err != nil {
		return nil, fmt.Errorf("error loading block entities of chunk %v: %w", pos, err)
	}
	w.loadIntoBlocks(data, blockEntities)
	return data, nil
}

// calculateLight calculates the light in the chunk passed and spreads the light of any of the surrounding
// neighbours if they have all chunks loaded around it as a result of the one passed.
func (w *World) calculateLight(c *chunk.Chunk, pos ChunkPos) {
	chunk.FillLight(c)

	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			// For all of the neighbours of this chunk, if they exist, check if all neighbours of that chunk
			// now exist because of this one.
			centrePos := ChunkPos{pos[0] + x, pos[1] + z}
			neighbour, ok := w.chunks[centrePos]
			if !ok {
				continue
			}
			neighbour.Lock()
			// We first attempt to spread the light of all neighbours into the ones around them.
			w.spreadLight(neighbour.Chunk, centrePos)
			neighbour.Unlock()
		}
	}
	// If the chunk loaded happened to be in the middle of a bunch of other chunks, we are able to spread it
	// right away, so we try to do that.
	w.spreadLight(c, pos)
}

// spreadLight spreads the light from the chunk passed at the position passed to all neighbours if each of
// them is loaded.
func (w *World) spreadLight(c *chunk.Chunk, pos ChunkPos) {
	neighbours, allPresent := make([]*chunk.Chunk, 0, 8), true
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			neighbour, ok := w.chunks[ChunkPos{pos[0] + x, pos[1] + z}]
			if !ok {
				allPresent = false
				break
			}
			if x != 0 || z != 0 {
				neighbours = append(neighbours, neighbour.Chunk)
			}
		}
	}
	if allPresent {
		for _, neighbour := range neighbours {
			neighbour.Lock()
		}
		// All neighbours of the current one are present, so we can spread the light from this chunk
		// to all neighbours.
		chunk.SpreadLight(c, neighbours)
		for _, neighbour := range neighbours {
			neighbour.Unlock()
		}
	}
}

// loadIntoBlocks loads the block entity data passed into blocks located in a specific chunk. The blocks that
// have block NBT will then be stored into memory.
func (w *World) loadIntoBlocks(c *chunkData, blockEntityData []map[string]interface{}) {
	c.e = make(map[cube.Pos]Block, len(blockEntityData))
	for _, data := range blockEntityData {
		pos := blockPosFromNBT(data)

		id := c.RuntimeID(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
		b, ok := BlockByRuntimeID(id)
		if !ok {
			w.log.Errorf("error loading block entity data: could not find block state by runtime ID %v", id)
			continue
		}
		if nbt, ok := b.(NBTer); ok {
			b = nbt.DecodeNBT(data).(Block)
		}
		c.e[pos] = b
	}
}

// saveChunk is called when a chunk is removed from the cache. We first compact the chunk, then we write it to
// the provider.
func (w *World) saveChunk(pos ChunkPos, c *chunkData) {
	c.Lock()
	// We allocate a new map for all block entities.
	m := make([]map[string]interface{}, 0, len(c.e))
	for pos, b := range c.e {
		if n, ok := b.(NBTer); ok {
			// Encode the block entities and add the 'x', 'y' and 'z' tags to it.
			data := n.EncodeNBT()
			data["x"], data["y"], data["z"] = int32(pos[0]), int32(pos[1]), int32(pos[2])
			m = append(m, data)
		}
	}
	if !w.rdonly.Load() {
		c.Compact()
		if err := w.provider().SaveChunk(pos, c.Chunk); err != nil {
			w.log.Errorf("error saving chunk %v to provider: %v", pos, err)
		}
		s := make([]SaveableEntity, 0, len(c.entities))
		for _, e := range c.entities {
			if saveable, ok := e.(SaveableEntity); ok {
				s = append(s, saveable)
			}
		}
		if err := w.provider().SaveEntities(pos, s); err != nil {
			w.log.Errorf("error saving entities in chunk %v to provider: %v", pos, err)
		}
		if err := w.provider().SaveBlockNBT(pos, m); err != nil {
			w.log.Errorf("error saving block NBT in chunk %v to provider: %v", pos, err)
		}
	}
	ent := c.entities
	c.entities = nil
	c.Unlock()

	for _, e := range ent {
		_ = e.Close()
	}
}

// initChunkCache initialises the chunk cache of the world to its default values.
func (w *World) initChunkCache() {
	w.chunkMu.Lock()
	w.chunks = make(map[ChunkPos]*chunkData)
	w.chunkMu.Unlock()
}

// chunkCacheJanitor runs until the world is running, cleaning chunks that are no longer in use from the cache.
func (w *World) chunkCacheJanitor() {
	t := time.NewTicker(time.Minute * 5)
	defer t.Stop()

	w.running.Add(1)
	chunksToRemove := map[ChunkPos]*chunkData{}
	for {
		select {
		case <-t.C:
			w.chunkMu.Lock()
			for pos, c := range w.chunks {
				c.Lock()
				v := len(c.v)
				c.Unlock()
				if v == 0 {
					chunksToRemove[pos] = c
					delete(w.chunks, pos)
					if w.lastPos == pos {
						w.lastChunk = nil
					}
				}
			}
			w.chunkMu.Unlock()

			for pos, c := range chunksToRemove {
				w.saveChunk(pos, c)
				delete(chunksToRemove, pos)
			}
		case <-w.closing:
			w.running.Done()
			return
		}
	}
}

// chunkData represents the data of a chunk including the block entities and viewers. This data is protected
// by the mutex present in the chunk.Chunk held.
type chunkData struct {
	*chunk.Chunk
	e        map[cube.Pos]Block
	v        []Viewer
	entities []Entity
}

// newChunkData returns a new chunkData wrapper around the chunk.Chunk passed.
func newChunkData(c *chunk.Chunk) *chunkData {
	return &chunkData{Chunk: c, e: map[cube.Pos]Block{}}
}
