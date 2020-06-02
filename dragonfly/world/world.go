package world

import (
	"context"
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/entity/physics"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/chunk"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/gamemode"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sirupsen/logrus"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
// World generally provides a synchronised state: All entities, blocks and players usually operate in this
// world, so World ensures that all its methods will always be safe for simultaneous calls.
type World struct {
	name string
	log  *logrus.Logger

	stopTick    context.Context
	cancelTick  context.CancelFunc
	doneTicking chan struct{}

	time        *int64
	timeStopped *uint32

	gameModeMu      sync.RWMutex
	defaultGameMode gamemode.GameMode

	blockMu      sync.RWMutex
	entityBlocks map[ChunkPos]map[BlockPos]Block

	handlerMu sync.RWMutex
	hand      Handler

	providerMu sync.RWMutex
	prov       Provider

	chunkMu sync.RWMutex
	// cCache holds a cache of chunks currently loaded. These chunks are cleared from this map after a while
	// of not being used.
	cCache map[ChunkPos]*chunk.Chunk
	// cTimes holds the times since last usage of the chunks at the same index in the map above.
	cTimes map[ChunkPos]time.Time

	genMu sync.RWMutex
	gen   Generator

	entityMu sync.RWMutex
	entities map[ChunkPos][]Entity

	viewerMu sync.RWMutex
	viewers  map[ChunkPos][]Viewer

	rdonly bool

	r         *rand.Rand
	simDistSq int32

	randomTickSpeed *uint32

	updateMu sync.Mutex
	// blockUpdates is a map of unix nano time values indexed by the block position at which an update is
	// scheduled. If the current time exceeds the unix time value passed, the block update will be performed
	// and the entry will be removed from the map.
	blockUpdates    map[BlockPos]int64
	updatePositions []BlockPos

	toTick []toTick

	chunkLoadMu sync.Mutex
}

// New creates a new initialised world. The world may be used right away, but it will not be saved or loaded
// from files until it has been given a different provider than the default. (NoIOProvider)
// By default, the name of the world will be 'World'.
func New(log *logrus.Logger, simulationDistance int) *World {
	randomTickSpeed := uint32(3)
	ctx, cancel := context.WithCancel(context.Background())
	w := &World{
		r:               rand.New(rand.NewSource(time.Now().Unix())),
		viewers:         map[ChunkPos][]Viewer{},
		entities:        map[ChunkPos][]Entity{},
		entityBlocks:    map[ChunkPos]map[BlockPos]Block{},
		blockUpdates:    map[BlockPos]int64{},
		defaultGameMode: gamemode.Survival{},
		prov:            NoIOProvider{},
		gen:             NopGenerator{},
		doneTicking:     make(chan struct{}),
		time:            new(int64),
		timeStopped:     new(uint32),
		simDistSq:       int32(simulationDistance * simulationDistance),
		randomTickSpeed: &randomTickSpeed,
		log:             log,
		stopTick:        ctx,
		cancelTick:      cancel,
		name:            "World",
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
	w.providerMu.RLock()
	n := w.name
	w.providerMu.RUnlock()
	return n
}

// Block reads a block from the position passed. If a chunk is not yet loaded at that position, the chunk is
// loaded, or generated if it could not be found in the world save, and the block returned. Chunks will be
// loaded synchronously.
func (w *World) Block(pos BlockPos) Block {
	if pos.OutOfBounds() {
		// Fast way out.
		return air()
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos), true)
	if err != nil {
		return air()
	}
	b, err := w.block(c, pos)
	if err != nil {
		w.log.Errorf("error getting block: %v", err)
		c.RUnlock()
		return air()
	}
	c.RUnlock()
	return b
}

// block reads a block from the world at the position passed. The block is assumed to be in the chunk passed,
// which is also assumed to be locked already or otherwise not yet accessible.
func (w *World) block(c *chunk.Chunk, pos BlockPos) (Block, error) {
	if pos.OutOfBounds() {
		// Fast way out.
		return air(), nil
	}
	id := c.RuntimeID(uint8(pos[0]&15), uint8(pos[1]), uint8(pos[2]&15), 0)

	state, ok := blockByRuntimeID(id)
	if !ok {
		// This should never happen.
		return nil, fmt.Errorf("could not find block state by runtime ID %v", id)
	}
	if _, ok := state.(NBTer); ok {
		// The block was also a block entity, so we look it up in the block entity map.
		w.blockMu.RLock()
		b, ok := w.entityBlocks[chunkPosFromBlockPos(pos)][pos]
		w.blockMu.RUnlock()
		if ok {
			return b, nil
		}
	}
	return state, nil
}

// runtimeID gets the block runtime ID at a specific position in the world.
//noinspection GoUnusedFunction
func runtimeID(w *World, pos BlockPos) uint32 {
	if pos[1] < 0 || pos[1] > 255 {
		// Fast way out.
		return 0
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos), true)
	if err != nil {
		return 0
	}
	rid := c.RuntimeID(uint8(pos[0]&0xf), uint8(pos[1]), uint8(pos[2]&0xf), 0)
	c.RUnlock()

	return rid
}

// SetBlock writes a block to the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
// SetBlock panics if the block passed has not yet been registered using RegisterBlock().
// Nil may be passed as the block to set the block to air.
// SetBlock should be avoided in situations where performance is critical when needing to set a lot of blocks
// to the world. BuildStructure may be used instead.
func (w *World) SetBlock(pos BlockPos, b Block) {
	if pos.OutOfBounds() {
		// Fast way out.
		return
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos), false)
	if err != nil {
		return
	}
	if err := w.setBlock(c, pos, b); err != nil {
		w.log.Errorf("error setting block: %v", err)
	}
	c.Unlock()
}

// setBlock sets a block at a position in a chunk to a given block. It does not lock the chunk passed, and
// assumes that is already done or that the chunk is otherwise inaccessible.
// Nil may be passed as the block to set the block to air.
func (w *World) setBlock(c *chunk.Chunk, pos BlockPos, b Block) error {
	w.blockMu.Lock()
	err := w.setBlockSilent(c, pos, b)
	w.blockMu.Unlock()
	for _, viewer := range w.Viewers(pos.Vec3()) {
		viewer.ViewBlockUpdate(pos, b)
	}
	return err
}

// setBlockSilent sets a block in the chunk passed at a specific position. Unlike setBlock, setBlockSilent
// does not send block updates to viewer.
// Callers of setBlockSilent must ensure that w.blockMu is locked while this method is called.
func (w *World) setBlockSilent(c *chunk.Chunk, pos BlockPos, b Block) error {
	runtimeID, ok := BlockRuntimeID(b)
	if !ok {
		return fmt.Errorf("runtime ID of block state %+v not found", b)
	}
	c.SetRuntimeID(uint8(pos[0]&15), uint8(pos[1]), uint8(pos[2]&15), 0, runtimeID)

	nbt, hasBlockEntity := b.(NBTer)
	if hasBlockEntity {
		// Encode the NBT of the block and add the 'x', 'y' and 'z' tags to it before saving it to the
		// chunk.
		data := nbt.EncodeNBT()
		data["x"], data["y"], data["z"] = int32(pos[0]), int32(pos[1]), int32(pos[2])
		c.SetBlockNBT(pos, data)

		chunkPos := chunkPosFromBlockPos(pos)
		if w.entityBlocks[chunkPos] == nil {
			w.entityBlocks[chunkPos] = map[BlockPos]Block{}
		}
		w.entityBlocks[chunkPos][pos] = b
	} else {
		// Clear any block NBT that might be present at the location.
		c.SetBlockNBT(pos, nil)
		delete(w.entityBlocks[chunkPosFromBlockPos(pos)], pos)
	}
	return nil
}

// breakParticle has its value set in the block_internal package.
var breakParticle func(b Block) Particle

// BreakBlock breaks a block at the position passed. Unlike when setting the block at that position to air,
// BreakBlock will also show particles.
func (w *World) BreakBlock(pos BlockPos) {
	old := w.Block(pos)
	w.SetBlock(pos, nil)
	w.AddParticle(pos.Vec3Centre(), breakParticle(old))
	w.doBlockUpdatesAround(pos)
}

// PlaceBlock places a block at the position passed. Unlike when using SetBlock, PlaceBlock also schedules
// block updates around the position.
func (w *World) PlaceBlock(pos BlockPos, b Block) {
	w.SetBlock(pos, b)
	w.doBlockUpdatesAround(pos)
}

// BuildStructure builds a Structure passed at a specific position in the world. Unlike SetBlock, it takes a
// Structure implementation, which provides blocks to be placed at a specific location.
// BuildStructure is specifically tinkered to be able to process a large batch of chunks simultaneously and
// will do so within much less time than separate SetBlock calls would.
// The method operates on a per-chunk basis, setting all blocks within a single chunk part of the structure
// before moving on to the next chunk.
func (w *World) BuildStructure(pos BlockPos, s Structure) {
	dim := s.Dimensions()
	width, height, length := dim[0], dim[1], dim[2]
	maxX, maxZ := pos[0]+width, pos[2]+length

	w.blockMu.Lock()
	for chunkX := pos[0] >> 4; chunkX < ((pos[0]+width)>>4)+1; chunkX++ {
		for chunkZ := pos[2] >> 4; chunkZ < ((pos[2]+length)>>4)+1; chunkZ++ {
			// We approach this on a per-chunk basis, so that we can keep only one chunk in memory at a time
			// while not needing to acquire a new chunk lock for every block. This also allows us not to send
			// block updates, but instead send a single chunk update once.

			chunkPos := ChunkPos{int32(chunkX), int32(chunkZ)}
			c, err := w.chunk(chunkPos, false)
			if err != nil {
				w.log.Errorf("error loading chunk for structure: %v", err)
			}
			f := func(x, y, z int) Block {
				if x>>4 == chunkX && z>>4 == chunkZ {
					b, _ := w.block(c, BlockPos{x, y, z})
					return b
				}
				return w.Block(BlockPos{x, y, z})
			}

			baseX, baseZ := chunkX<<4, chunkZ<<4
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
					for y := 0; y < height; y++ {
						if y+pos[1] > 255 {
							// We've hit the height limit for blocks.
							break
						} else if y+pos[1] < 0 {
							// We've got a block below the minimum, but other blocks might still reach above
							// it, so don't break but continue.
							continue
						}
						placePos := BlockPos{xOffset, y + pos[1], zOffset}
						if b := s.At(xOffset-pos[0], y, zOffset-pos[2], f); b != nil {
							if err := w.setBlockSilent(c, placePos, b); err != nil {
								w.log.Errorf("error setting block of structure: %v", err)
							}
						}
					}
				}
			}
			// After setting all blocks of the structure within a single chunk, we show the new chunk to all
			// viewers once, and unlock it.
			for _, viewer := range w.chunkViewers(chunkPos) {
				viewer.ViewChunk(chunkPos, c)
			}
			c.Unlock()
		}
	}
	w.blockMu.Unlock()
}

// Light returns the light level at the position passed. This is the highest of the sky and block light.
// The light value returned is a value in the range 0-15, where 0 means there is no light present, whereas
// 15 means the block is fully lit.
func (w *World) Light(pos BlockPos) uint8 {
	if pos[1] < 0 || pos[1] > 255 {
		// Fast way out.
		return 0
	}
	c, err := w.chunk(chunkPosFromBlockPos(pos), true)
	if err != nil {
		return 0
	}
	l := c.Light(uint8(pos[0]), uint8(pos[1]), uint8(pos[2]))
	c.RUnlock()

	return l
}

// Time returns the current time of the world. The time is incremented every 1/20th of a second, unless
// World.StopTime() is called.
func (w *World) Time() int {
	return int(atomic.LoadInt64(w.time))
}

// SetTime sets the new time of the world. SetTime will always work, regardless of whether the time is stopped
// or not.
func (w *World) SetTime(new int) {
	atomic.StoreInt64(w.time, int64(new))
	for _, viewer := range w.allViewers() {
		viewer.ViewTime(new)
	}
}

// StopTime stops the time in the world. When called, the time will no longer cycle and the world will remain
// at the time when StopTime is called. The time may be restarted by calling World.StartTime().
// StopTime will not do anything if the time is already stopped.
func (w *World) StopTime() {
	atomic.StoreUint32(w.timeStopped, 1)
}

// StartTime restarts the time in the world. When called, the time will start cycling again and the day/night
// cycle will continue. The time may be stopped again by calling World.StopTime().
// StartTime will not do anything if the time is already started.
func (w *World) StartTime() {
	atomic.StoreUint32(w.timeStopped, 0)
}

// AddParticle spawns a particle at a given position in the world. Viewers that are viewing the chunk will be
// shown the particle.
func (w *World) AddParticle(pos mgl64.Vec3, p Particle) {
	p.Spawn(w, pos)
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewParticle(pos, p)
	}
}

// PlaySound plays a sound at a specific position in the world. Viewers of that position will be able to hear
// the sound if they're close enough.
func (w *World) PlaySound(pos mgl64.Vec3, s sound.Sound) {
	for _, viewer := range w.Viewers(pos) {
		viewer.ViewSound(pos, s)
	}
}

// entityWorlds holds a list of all entities added to a world. It may be used to lookup the world that an
// entity is currently in.
var entityWorlds = map[Entity]*World{}
var worldsMu sync.RWMutex

// AddEntity adds an entity to the world at the position that the entity has. The entity will be visible to
// all viewers of the world that have the chunk of the entity loaded.
// If the chunk that the entity is in is not yet loaded, it will first be loaded.
// If the entity passed to AddEntity is currently in a world, it is first removed from that world.
func (w *World) AddEntity(e Entity) {
	if e.World() != nil {
		e.World().RemoveEntity(e)
	}
	chunkPos := chunkPosFromVec3(e.Position())
	c, err := w.chunk(chunkPos, true)
	if err != nil {
		w.log.Errorf("error loading chunk to add entity: %v", err)
	}
	c.RUnlock()

	worldsMu.Lock()
	entityWorlds[e] = w
	worldsMu.Unlock()

	w.entityMu.Lock()
	w.entities[chunkPos] = append(w.entities[chunkPos], e)
	w.entityMu.Unlock()

	w.viewerMu.RLock()
	for _, viewer := range w.viewers[chunkPos] {
		// We show the entity to all viewers currently in the chunk that the entity is spawned in.
		showEntity(e, viewer)
	}
	w.viewerMu.RUnlock()
}

// RemoveEntity removes an entity from the world that is currently present in it. Any viewers of the entity
// will no longer be able to see it.
// RemoveEntity operates assuming the position of the entity is the same as where it is currently in the
// world. If it can not find it there, it will loop through all entities and try to find it.
// RemoveEntity assumes the entity is currently loaded and in a loaded chunk. If not, the function will not do
// anything.
func (w *World) RemoveEntity(e Entity) {
	chunkPos := chunkPosFromVec3(e.Position())

	worldsMu.Lock()
	delete(entityWorlds, e)
	worldsMu.Unlock()

	w.entityMu.Lock()
	if _, ok := w.chunkFromCache(chunkPos); !ok {
		// The chunk wasn't loaded, so we can't remove any entity from the chunk.
		w.entityMu.Unlock()
		return
	}
	if !w.removeEntity(chunkPos, e) {
		w.log.Debugf("failed removing entity %T{%v} at chunk position %v", e, e, chunkPos)
	}
	w.entityMu.Unlock()
}

// EntitiesWithin does a lookup through the entities in the chunks touched by the AABB passed, returning all
// those which are contained within the AABB when it comes to their position.
func (w *World) EntitiesWithin(aabb physics.AABB) []Entity {
	// Make an estimate of 16 entities on average.
	m := make([]Entity, 0, 16)

	// We expand it by 3 blocks in all horizontal directions to account for entities that may be in
	// neighbouring chunks while having a bounding box that extends into the current one.
	minPos, maxPos := chunkPosFromVec3(aabb.Min()), chunkPosFromVec3(aabb.Max())

	w.entityMu.RLock()
	for x := minPos[0]; x <= maxPos[0]; x++ {
		for z := minPos[1]; z <= maxPos[1]; z++ {
			chunkEntities, ok := w.entities[ChunkPos{x, z}]
			if !ok {
				// Chunk wasn't currently loaded or had no entities in it, so we can continue with the next.
				continue
			}
			for _, entity := range chunkEntities {
				if aabb.Vec3Within(entity.Position()) {
					// The entity position was within the AABB, so we add it to the slice to return.
					m = append(m, entity)
				}
			}
		}
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
func (w *World) Spawn() BlockPos {
	return w.provider().WorldSpawn()
}

// SetSpawn sets the spawn of the world to a different position. The player will be spawned in the center of
// this position when newly joining.
func (w *World) SetSpawn(pos BlockPos) {
	w.provider().SetWorldSpawn(pos)
}

// DefaultGameMode returns the default game mode of the world. When players join, they are given this game
// mode.
// The default game mode may be changed using SetDefaultGameMode().
func (w *World) DefaultGameMode() gamemode.GameMode {
	w.gameModeMu.RLock()
	defer w.gameModeMu.RUnlock()
	return w.defaultGameMode
}

// SetDefaultGameMode changes the default game mode of the world. When players join, they are then given that
// game mode.
func (w *World) SetDefaultGameMode(mode gamemode.GameMode) {
	w.gameModeMu.Lock()
	w.defaultGameMode = mode
	w.gameModeMu.Unlock()
}

// SetRandomTickSpeed sets the random tick speed of blocks. By default, each sub chunk has 3 blocks randomly
// ticked per sub chunk, so the default value is 3. Setting this value to 0 will stop random ticking
// altogether, while setting it higher results in faster ticking.
func (w *World) SetRandomTickSpeed(v int) {
	atomic.StoreUint32(w.randomTickSpeed, uint32(v))
}

// ScheduleBlockUpdate schedules a block update at the position passed after a specific delay. If the block at
// that position does not handle block updates, nothing will happen.
func (w *World) ScheduleBlockUpdate(pos BlockPos, delay time.Duration) {
	if pos.OutOfBounds() {
		return
	}
	w.updateMu.Lock()
	if _, exists := w.blockUpdates[pos]; exists {
		w.updateMu.Unlock()
		return
	}
	w.blockUpdates[pos] = time.Now().Add(delay).UnixNano()
	w.updateMu.Unlock()
}

// doBlockUpdatesAround schedules block updates directly around and on the position passed.
func (w *World) doBlockUpdatesAround(pos BlockPos) {
	if pos.OutOfBounds() {
		return
	}

	changed := pos
	w.updateNeighbour(pos, changed)
	pos.Neighbours(func(pos BlockPos) {
		w.updateNeighbour(pos, changed)
	})
}

// updateNeighbour ticks the position passed as a result of the neighbour passed being updated.
func (w *World) updateNeighbour(pos, changedNeighbour BlockPos) {
	if ticker, ok := w.Block(pos).(NeighbourUpdateTicker); ok {
		ticker.NeighbourUpdateTick(pos, changedNeighbour, w)
	}
}

// Provider changes the provider of the world to the provider passed. If nil is passed, the NoIOProvider
// will be set, which does not read or write any data.
func (w *World) Provider(p Provider) {
	w.providerMu.Lock()
	defer w.providerMu.Unlock()

	if p == nil {
		p = NoIOProvider{}
	}
	w.prov = p
	w.name = p.WorldName()
	w.gameModeMu.Lock()
	w.defaultGameMode = p.LoadDefaultGameMode()
	w.gameModeMu.Unlock()
	atomic.StoreInt64(w.time, p.LoadTime())
	if timeRunning := p.LoadTimeCycle(); !timeRunning {
		atomic.StoreUint32(w.timeStopped, 1)
	}
	w.initChunkCache()
}

// ReadOnly makes the world read only. Chunks will no longer be saved to disk, just like entities and data
// in the level.dat.
func (w *World) ReadOnly() {
	w.rdonly = true
}

// Generator changes the generator of the world to the one passed. If nil is passed, the generator is set to
// the default, NopGenerator.
func (w *World) Generator(g Generator) {
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
	w.handlerMu.Lock()
	defer w.handlerMu.Unlock()

	if h == nil {
		h = NopHandler{}
	}
	w.hand = h
}

// Viewers returns a list of all viewers viewing the position passed. A viewer will be assumed to be watching
// if the position is within one of the chunks that the viewer is watching.
func (w *World) Viewers(pos mgl64.Vec3) []Viewer {
	return w.chunkViewers(chunkPosFromVec3(pos))
}

// Close closes the world and saves all chunks currently loaded.
func (w *World) Close() error {
	w.cancelTick()
	<-w.doneTicking

	w.viewerMu.Lock()
	w.viewers = map[ChunkPos][]Viewer{}
	w.viewerMu.Unlock()

	w.log.Debug("Saving chunks in memory to disk...")

	w.chunkMu.Lock()
	chunksToSave := make(map[ChunkPos]*chunk.Chunk, len(w.cCache))
	for pos, c := range w.cCache {
		// We delete all chunks from the cache and save them to the provider.
		delete(w.cTimes, pos)
		delete(w.cCache, pos)
		chunksToSave[pos] = c
	}
	w.chunkMu.Unlock()

	for pos, c := range chunksToSave {
		w.saveChunk(pos, c)
	}

	if !w.rdonly {
		w.log.Debug("Updating level.dat values...")
		w.provider().SaveTime(atomic.LoadInt64(w.time))
		w.provider().SaveTimeCycle(atomic.LoadUint32(w.timeStopped) == 0)

		w.gameModeMu.RLock()
		w.provider().SaveDefaultGameMode(w.defaultGameMode)
		w.gameModeMu.RUnlock()
	}

	w.log.Debug("Closing provider...")
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

	tick := 0
	for {
		select {
		case <-ticker.C:
			w.tick(tick)
			tick++
		case <-w.stopTick.Done():
			// The world was closed, so we should stop ticking.
			w.doneTicking <- struct{}{}
			return
		}
	}
}

// tick ticks the world and updates the time, blocks and entities that require updates.
func (w *World) tick(tick int) {
	if atomic.LoadUint32(w.timeStopped) == 0 {
		// Only if the time is not stopped, we add one to the current time.
		atomic.AddInt64(w.time, 1)
	}
	if tick%20 == 0 {
		for _, viewer := range w.allViewers() {
			viewer.ViewTime(int(atomic.LoadInt64(w.time)))
		}
	}
	w.tickEntities()
	w.tickRandomBlocks()
	w.tickScheduledBlocks()
}

// tickScheduledBlocks executes scheduled block ticks in chunks that are still loaded at the time of
// execution.
func (w *World) tickScheduledBlocks() {
	currentNano := time.Now().UnixNano()

	w.updateMu.Lock()
	for pos, unixNano := range w.blockUpdates {
		if unixNano <= currentNano {
			w.updatePositions = append(w.updatePositions, pos)
			delete(w.blockUpdates, pos)
		}
	}
	w.updateMu.Unlock()

	for _, pos := range w.updatePositions {
		if ticker, ok := w.Block(pos).(ScheduledTicker); ok {
			ticker.ScheduledTick(pos, w)
		}
	}

	w.updatePositions = w.updatePositions[:0]
}

// toTick is a struct used to keep track of blocks that need to be ticked upon a random tick.
type toTick struct {
	b   RandomTicker
	pos BlockPos
}

// tickRandomBlocks executes random block ticks in each sub chunk in the world that has at least one viewer
// registered.
func (w *World) tickRandomBlocks() {
	if w.simDistSq == 0 {
		// NOP if the simulation distance is 0.
		return
	}
	viewers := w.allViewers()

	w.chunkMu.RLock()
	for pos, c := range w.cCache {
		withinSimDist := false
		for _, viewer := range viewers {
			chunkPos := chunkPosFromVec3(viewer.Position())
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

		c.RLock()
		// In total we generate 3 random blocks per sub chunk.
		for j := uint32(0); j < atomic.LoadUint32(w.randomTickSpeed); j++ {
			// We generate 3 random uint64s. Out of a single uint64, we can pull 16 uint4s, which means we can
			// obtain a total of 16 coordinates on one axis from one uint64. One for each sub chunk.
			ra, rb, rc := int(w.r.Uint64()), int(w.r.Uint64()), int(w.r.Uint64())
			for i := 0; i < 64; i += 4 {
				if !c.SubChunkPresent(uint8(i >> 2)) {
					// No sub chunk present, so skip it right away.
					continue
				}
				x, y, z := ra>>i&0xf, (rb>>i&0xf)+i<<2, rc>>i&0xf

				blockPos := pos.BlockPos().Add(BlockPos{x, y, z})

				// Generally we would want to make sure the block has its block entities, but provided blocks
				// with block entities are generally ticked already, we are safe to assume that blocks
				// implementing the RandomTicker don't rely on additional block entity data.
				rid := c.RuntimeID(uint8(blockPos[0]&15), uint8(blockPos[1]), uint8(blockPos[2]&15), 0)
				if rid == 0 {
					// The block was air, take the fast route out.
					continue
				}
				b, _ := blockByRuntimeID(rid)

				if randomTicker, ok := b.(RandomTicker); ok {
					w.toTick = append(w.toTick, toTick{b: randomTicker, pos: blockPos})
				}
			}
		}
		c.RUnlock()
	}
	w.chunkMu.RUnlock()

	for _, a := range w.toTick {
		a.b.RandomTick(a.pos, w, w.r)
	}
	w.toTick = w.toTick[:0]
}

// tickEntities ticks all entities in the world, making sure they are still located in the correct chunks and
// updating where necessary.
func (w *World) tickEntities() {
	w.entityMu.Lock()
	entitiesToTick := make([]TickerEntity, 0, len(w.entities)*8)
	for chunkPos, entities := range w.entities {
		chunkEntities := make([]Entity, 0, len(entities))
		for _, entity := range entities {
			if ticker, ok := entity.(TickerEntity); ok {
				entitiesToTick = append(entitiesToTick, ticker)
			}

			// The entity was stored using an outdated chunk position. We update it and make sure it is ready
			// for viewers to view it.
			newChunkPos := chunkPosFromVec3(entity.Position())
			if newChunkPos != chunkPos {
				w.entities[newChunkPos] = append(w.entities[newChunkPos], entity)

				w.viewerMu.RLock()
				for _, viewer := range w.viewers[chunkPos] {
					if !w.hasViewer(newChunkPos, viewer) {
						// First we hide the entity from all viewers that were previously viewing it, but no
						// longer are.
						viewer.HideEntity(entity)
					}
				}
				for _, viewer := range w.viewers[newChunkPos] {
					if !w.hasViewer(chunkPos, viewer) {
						// Then we show the entity to all viewers that are now viewing the entity in the new
						// chunk.
						showEntity(entity, viewer)
					}
				}
				w.viewerMu.RUnlock()
				continue
			}
			chunkEntities = append(chunkEntities, entity)
		}
		if len(chunkEntities) == 0 {
			// There are no more entities stored in this chunk: We delete the entry from the map.
			delete(w.entities, chunkPos)
			continue
		}
		w.entities[chunkPos] = chunkEntities
	}
	w.entityMu.Unlock()

	for _, ticker := range entitiesToTick {
		if _, ok := OfEntity(ticker.(Entity)); !ok {
			continue
		}
		// We gather entities to tick and tick them later, so that the lock on the entity mutex is no longer
		// active.
		ticker.Tick()
	}
}

// removeEntity attempts to remove an entity located in a chunk at the chunk position passed. If found, it
// removes the entity and returns true. If it can't be found, removeEntity returns false.
func (w *World) removeEntity(chunkPos ChunkPos, e Entity) (found bool) {
	n := make([]Entity, 0, len(w.entities[chunkPos]))
	for _, entity := range w.entities[chunkPos] {
		if entity != e {
			n = append(n, entity)
			continue
		}
		w.viewerMu.RLock()
		for _, viewer := range w.viewers[chunkPos] {
			viewer.HideEntity(e)
		}
		w.viewerMu.RUnlock()
		found = true
	}
	if len(n) == 0 {
		// The entity is the last in the chunk, so we can delete the value from the map.
		delete(w.entities, chunkPos)
		return
	}
	w.entities[chunkPos] = n
	return
}

// addViewer adds a viewer to the world at a given position. Any events that happen in the chunk at that
// position, such as block changes, entity changes etc., will be sent to the viewer.
func (w *World) addViewer(pos ChunkPos, viewer Viewer) {
	w.viewerMu.Lock()
	w.viewers[pos] = append(w.viewers[pos], viewer)
	w.viewerMu.Unlock()

	// After adding the viewer to the chunk, we also need to send all entities currently in the chunk that the
	// viewer is added to.
	w.entityMu.RLock()
	for _, entity := range w.entities[pos] {
		showEntity(entity, viewer)
	}
	w.entityMu.RUnlock()

	viewer.ViewTime(w.Time())
}

// removeViewer removes a viewer from the world at a given position. All entities will be hidden from the
// viewer and no more calls will be made when events in the chunk happen.
func (w *World) removeViewer(pos ChunkPos, viewer Viewer) {
	w.viewerMu.Lock()
	n := make([]Viewer, 0, len(w.viewers[pos]))
	for _, v := range w.viewers[pos] {
		if v != viewer {
			// Add all viewers but the one to remove to the new viewers slice.
			n = append(n, v)
		}
	}
	if len(n) == 0 {
		delete(w.viewers, pos)
	} else {
		w.viewers[pos] = n
	}
	w.viewerMu.Unlock()

	// After removing the viewer from the chunk, we also need to hide all entities from the viewer.
	w.entityMu.RLock()
	for _, entity := range w.entities[pos] {
		viewer.HideEntity(entity)
	}
	w.entityMu.RUnlock()
}

// hasViewer checks if a chunk at a particular chunk position has the viewer passed. If so, true is returned.
func (w *World) hasViewer(pos ChunkPos, viewer Viewer) bool {
	for _, v := range w.viewers[pos] {
		if v == viewer {
			return true
		}
	}
	return false
}

// allViewers returns a list of all viewers of the world, regardless of where in the world they are viewing.
func (w *World) allViewers() []Viewer {
	var v []Viewer
	found := make(map[Viewer]struct{})
	w.viewerMu.RLock()
	for _, c := range w.viewers {
		for _, viewer := range c {
			if _, ok := found[viewer]; ok {
				// We've already found this viewer in another chunk. Don't add it again.
				continue
			}
			found[viewer] = struct{}{}
			v = append(v, viewer)
		}
	}
	w.viewerMu.RUnlock()
	return v
}

// provider returns the provider of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) provider() Provider {
	w.providerMu.RLock()
	provider := w.prov
	w.providerMu.RUnlock()
	return provider
}

// Handler returns the Handler of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) Handler() Handler {
	w.handlerMu.RLock()
	handler := w.hand
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
func (w *World) chunkFromCache(pos ChunkPos) (*chunk.Chunk, bool) {
	w.chunkMu.RLock()
	c, ok := w.cCache[pos]
	w.chunkMu.RUnlock()
	return c, ok
}

// storeChunkToCache stores a chunk at a position passed to the chunk cache.
func (w *World) storeChunkToCache(pos ChunkPos, c *chunk.Chunk) {
	w.chunkMu.Lock()
	w.cCache[pos] = c
	w.cTimes[pos] = time.Now().Add(time.Minute * 5)
	w.chunkMu.Unlock()
}

// chunkViewers returns a list of all viewers of a chunk at a given position.
func (w *World) chunkViewers(pos ChunkPos) []Viewer {
	w.viewerMu.RLock()
	viewers := make([]Viewer, len(w.viewers[pos]))
	copy(viewers, w.viewers[pos])
	w.viewerMu.RUnlock()
	return viewers
}

// showEntity shows an entity to a viewer of the world. It makes sure everything of the entity, including the
// items held, is shown.
func showEntity(e Entity, viewer Viewer) {
	viewer.ViewEntity(e)
	viewer.ViewEntityState(e, e.State())
	viewer.ViewEntityItems(e)
	viewer.ViewEntityArmour(e)
}

// chunk reads a chunk from the position passed. If a chunk at that position is not yet loaded, the chunk is
// loaded from the provider, or generated if it did not yet exist. Both of these actions are done
// synchronously.
// An error is returned if the chunk could not be loaded successfully.
// chunk locks the chunk returned, meaning that any call to chunk made at the same time has to wait until the
// user calls Chunk.Unlock() on the chunk returned.
func (w *World) chunk(pos ChunkPos, readOnly bool) (*chunk.Chunk, error) {
	var needsLight bool
	var err error

	w.chunkLoadMu.Lock()
	c, ok := w.chunkFromCache(pos)
	if !ok {
		c, err = w.loadChunk(pos)
		if err != nil {
			return nil, err
		}
		w.storeChunkToCache(pos, c)
		needsLight = true
	}
	w.chunkLoadMu.Unlock()

	if needsLight {
		w.calculateLight(c, pos)
	}

	// Update the timestamp to that it doesn't expire after we just used it.
	w.storeChunkToCache(pos, c)

	if readOnly {
		c.RLock()
	} else {
		c.Lock()
	}
	return c, nil
}

// loadChunk attempts to load a chunk from the provider, or generates a chunk if one doesn't currently exist.
func (w *World) loadChunk(pos ChunkPos) (c *chunk.Chunk, err error) {
	var found bool
	c, found, err = w.provider().LoadChunk(pos)
	if err != nil {
		return nil, fmt.Errorf("error loading chunk %v: %w", pos, err)
	}
	if !found {
		// The provider doesn't have a chunk saved at this position, so we generate a new one.
		c = chunk.New()
		w.generator().GenerateChunk(pos, c)
	} else {
		entities, err := w.provider().LoadEntities(pos)
		if err != nil {
			return nil, fmt.Errorf("error loading entities of chunk %v: %w", pos, err)
		}
		if len(entities) != 0 {
			for _, e := range entities {
				w.AddEntity(e)
			}
		}
		blockEntities, err := w.provider().LoadBlockNBT(pos)
		if err != nil {
			return nil, fmt.Errorf("error loading block entities of chunk %v: %w", pos, err)
		}
		w.loadIntoBlocks(c, blockEntities)
	}
	return c, nil
}

// calculateLight calculates the light in the chunk passed and spreads the light of any of the surrounding
// neighbours if they have all chunks loaded around it as a result of the one passed.
func (w *World) calculateLight(c *chunk.Chunk, pos ChunkPos) {
	c.Lock()
	chunk.FillLight(c)
	c.Unlock()

	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			// For all of the neighbours of this chunk, if they exist, check if all neighbours of that chunk
			// now exist because of this one.
			centrePos := ChunkPos{pos[0] + x, pos[1] + z}
			neighbour, ok := w.chunkFromCache(centrePos)
			if !ok {
				continue
			}
			neighbour.Lock()
			// We first attempt to spread the light of all neighbours into the ones around them.
			w.spreadLight(neighbour, centrePos)
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
			neighbour, ok := w.chunkFromCache(ChunkPos{pos[0] + x, pos[1] + z})
			if !ok {
				allPresent = false
				break
			}
			if !(x == 0 && z == 0) {
				neighbours = append(neighbours, neighbour)
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
func (w *World) loadIntoBlocks(c *chunk.Chunk, blockEntityData []map[string]interface{}) {
	for _, data := range blockEntityData {
		pos := blockPosFromNBT(data)
		b, err := w.block(c, pos)
		if err != nil {
			w.log.Errorf("error loading block for block entity: %v", err)
			continue
		}
		if nbt, ok := b.(NBTer); ok {
			b = nbt.DecodeNBT(data).(Block)
		}
		if err := w.setBlock(c, pos, b); err != nil {
			w.log.Errorf("error setting block with block entity back: %v", err)
		}
	}
}

// saveChunk is called when a chunk is removed from the cache. We first compact the chunk, then we write it to
// the provider.
func (w *World) saveChunk(pos ChunkPos, c *chunk.Chunk) {
	w.entityMu.Lock()
	entities := w.entities[pos]
	delete(w.entities, pos)
	w.entityMu.Unlock()

	w.blockMu.Lock()
	// We allocate a new map for all block entities.
	m := make(map[[3]int]map[string]interface{}, len(w.entityBlocks))
	for pos, b := range w.entityBlocks[pos] {
		// Encode the block entities and add the 'x', 'y' and 'z' tags to it.
		data := b.(NBTer).EncodeNBT()
		data["x"], data["y"], data["z"] = int32(pos[0]), int32(pos[1]), int32(pos[2])
		m[pos] = data
	}
	delete(w.entityBlocks, pos)
	w.blockMu.Unlock()

	if !w.rdonly {
		c.Lock()
		c.Compact()
		c.Unlock()
		if err := w.provider().SaveChunk(pos, c); err != nil {
			w.log.Errorf("error saving chunk %v to provider: %v", pos, err)
		}
		if err := w.provider().SaveEntities(pos, entities); err != nil {
			w.log.Errorf("error saving entities in chunk %v to provider: %v", pos, err)
		}
		if err := w.provider().SaveBlockNBT(pos, m); err != nil {
			w.log.Errorf("error saving block NBT in chunk %v to provider: %v", pos, err)
		}
	}

	for _, entity := range entities {
		_ = entity.Close()
	}
}

// initChunkCache initialises the chunk cache of the world to its default values.
func (w *World) initChunkCache() {
	w.chunkMu.Lock()
	w.cCache = make(map[ChunkPos]*chunk.Chunk)
	w.cTimes = make(map[ChunkPos]time.Time)
	w.chunkMu.Unlock()
}

// chunkCacheJanitor runs until the world is closed, cleaning chunks that are no longer in use from the cache.
func (w *World) chunkCacheJanitor() {
	t := time.NewTicker(time.Minute * 5)
	chunksToRemove := map[ChunkPos]*chunk.Chunk{}
	for {
		select {
		case <-t.C:
			w.chunkMu.Lock()
			for pos, t := range w.cTimes {
				if len(w.chunkViewers(pos)) != 0 {
					w.cTimes[pos] = time.Now().Add(time.Minute * 5)
					continue
				}
				if time.Until(t) <= 0 {
					chunksToRemove[pos] = w.cCache[pos]
					delete(w.cTimes, pos)
					delete(w.cCache, pos)
				}
			}
			w.chunkMu.Unlock()

			for pos, c := range chunksToRemove {
				w.saveChunk(pos, c)
			}
			for k := range chunksToRemove {
				delete(chunksToRemove, k)
			}
		case <-w.stopTick.Done():
			return
		}
	}
}
