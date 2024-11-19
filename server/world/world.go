package world

import (
	"encoding/binary"
	"errors"
	"fmt"
	"iter"
	"math/rand"
	"sync"
	"time"

	"github.com/df-mc/goleveldb/leveldb"

	"slices"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"sync/atomic"
)

// World implements a Minecraft world. It manages all aspects of what players can see, such as blocks,
// entities and particles.
// World generally provides a synchronised state: All entities, blocks and players usually operate in this
// world, so World ensures that all its methods will always be safe for simultaneous calls.
// A nil *World is safe to use but not functional.
type World struct {
	conf Config
	ra   cube.Range

	queue chan transaction

	// advance is a bool that specifies if this World should advance the current
	// tick, time and weather saved in the Settings struct held by the World.
	advance bool

	o sync.Once

	set     *Settings
	handler atomic.Pointer[Handler]

	weather
	ticker

	closing chan struct{}
	running sync.WaitGroup

	// chunks holds a cache of chunks currently loaded. These chunks are cleared from this map after some time
	// of not being used.
	chunks map[ChunkPos]*Column

	// entities holds a map of entities currently loaded and the last ChunkPos that the Entity was in.
	// These are tracked so that a call to RemoveEntity can find the correct Entity.
	entities map[*EntityHandle]ChunkPos

	r *rand.Rand

	// scheduledUpdates is a map of tick time values indexed by the block position at which an update is
	// scheduled. If the current tick exceeds the tick value passed, the block update will be performed
	// and the entry will be removed from the map.
	scheduledUpdates map[cube.Pos]int64
	neighbourUpdates []neighbourUpdate

	viewers map[*Loader]Viewer
}

type transaction struct {
	c chan struct{}
	f func(tx *Tx)
}

// New creates a new initialised world. The world may be used right away, but it will not be saved or loaded
// from files until it has been given a different provider than the default. (NopProvider)
// By default, the name of the world will be 'World'.
func New() *World {
	var conf Config
	return conf.New()
}

// Name returns the display name of the world. Generally, this name is displayed at the top of the player list
// in the pause screen in-game.
// If a provider is set, the name will be updated according to the name that it provides.
func (w *World) Name() string {
	w.set.Lock()
	defer w.set.Unlock()
	return w.set.Name
}

// Dimension returns the Dimension assigned to the World in world.New. The sky colour and behaviour of a variety of
// world features differ based on the Dimension assigned to a World.
func (w *World) Dimension() Dimension {
	return w.conf.Dim
}

// Range returns the range in blocks of the World (min and max). It is equivalent to calling World.Dimension().Range().
func (w *World) Range() cube.Range {
	return w.ra
}

func (w *World) Exec(f func(tx *Tx)) <-chan struct{} {
	c := make(chan struct{}, 1)
	w.queue <- transaction{c: c, f: f}
	return c
}

func (w *World) handleTransactions() {
	w.running.Add(1)
	for {
		select {
		case queuedTx := <-w.queue:
			tx := &Tx{w: w}
			queuedTx.f(tx)
			close(queuedTx.c)
		case <-w.closing:
			w.running.Done()
			return
		}
	}
}

// EntityRegistry returns the EntityRegistry that was passed to the World's
// Config upon construction.
func (w *World) EntityRegistry() EntityRegistry {
	return w.conf.Entities
}

// Block reads a block from the position passed. If a chunk is not yet loaded at that position, the chunk is
// loaded, or generated if it could not be found in the world save, and the block returned. Chunks will be
// loaded synchronously.
func (w *World) block(pos cube.Pos) Block {
	return w.blockInChunk(w.chunk(chunkPosFromBlockPos(pos)), pos)
}

// blockInChunk reads a block from the world at the position passed. The block
// is assumed to be in the chunk passed, which is also assumed to be locked
// already or otherwise not yet accessible.
func (w *World) blockInChunk(c *Column, pos cube.Pos) Block {
	if pos.OutOfBounds(w.ra) {
		// Fast way out.
		return air()
	}
	rid := c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 0)
	if nbtBlocks[rid] {
		// The block was also a block entity, so we look it up in the map.
		if b, ok := c.BlockEntities[pos]; ok {
			return b
		}
		// Despite being a block with NBT, the block didn't actually have any
		// stored NBT yet. We add it here and update the block.
		nbtB := blockByRuntimeIDOrAir(rid).(NBTer).DecodeNBT(map[string]any{}).(Block)
		c.BlockEntities[pos] = nbtB
		for _, v := range c.viewers {
			v.ViewBlockUpdate(pos, nbtB, 0)
		}
		return nbtB
	}
	return blockByRuntimeIDOrAir(rid)
}

// Biome reads the biome at the position passed. If a chunk is not yet loaded at that position, the chunk is
// loaded, or generated if it could not be found in the world save, and the biome returned. Chunks will be
// loaded synchronously.
func (w *World) biome(pos cube.Pos) Biome {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return ocean()
	}
	id := int(w.chunk(chunkPosFromBlockPos(pos)).Biome(uint8(pos[0]), int16(pos[1]), uint8(pos[2])))
	b, ok := BiomeByID(id)
	if !ok {
		w.conf.Log.Error("biome not found by ID", "ID", id)
	}
	return b
}

// HighestLightBlocker gets the Y value of the highest fully light blocking block at the x and z values
// passed in the world.
func (w *World) highestLightBlocker(x, z int) int {
	return int(w.chunk(ChunkPos{int32(x >> 4), int32(z >> 4)}).HighestLightBlocker(uint8(x), uint8(z)))
}

// HighestBlock looks up the highest non-air block in the world at a specific x and z in the world. The y
// value of the highest block is returned, or 0 if no blocks were present in the column.
func (w *World) highestBlock(x, z int) int {
	return int(w.chunk(ChunkPos{int32(x >> 4), int32(z >> 4)}).HighestBlock(uint8(x), uint8(z)))
}

// highestObstructingBlock returns the highest block in the world at a given x and z that has at least a solid top or
// bottom face.
func (w *World) highestObstructingBlock(x, z int) int {
	yHigh := w.highestBlock(x, z)
	for y := yHigh; y >= w.Range()[0]; y-- {
		pos := cube.Pos{x, y, z}
		m := w.block(pos).Model()
		// TODO: Work out how to pass a proper BlockSource here.
		if m.FaceSolid(pos, cube.FaceUp, nil) || m.FaceSolid(pos, cube.FaceDown, nil) {
			return y
		}
	}
	return w.Range()[0]
}

// SetOpts holds several parameters that may be set to disable updates in the World of different kinds as a result of
// a call to SetBlock.
type SetOpts struct {
	// DisableBlockUpdates makes SetBlock not update any neighbouring blocks as a result of the SetBlock call.
	DisableBlockUpdates bool
	// DisableLiquidDisplacement disables the displacement of liquid blocks to the second layer (or back to the first
	// layer, if it already was on the second layer). Disabling this is not strongly recommended unless performance is
	// very important or where it is known no liquid can be present anyway.
	DisableLiquidDisplacement bool
}

// SetBlock writes a block to the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
// SetBlock panics if the block passed has not yet been registered using RegisterBlock().
// Nil may be passed as the block to set the block to air.
//
// A SetOpts struct may be passed to additionally modify behaviour of SetBlock, specifically to improve performance
// under specific circumstances. Nil should be passed where performance is not essential, to make sure the world is
// updated adequately.
//
// SetBlock should be avoided in situations where performance is critical when needing to set a lot of blocks
// to the world. BuildStructure may be used instead.
func (w *World) setBlock(pos cube.Pos, b Block, opts *SetOpts) {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return
	}
	if opts == nil {
		opts = &SetOpts{}
	}

	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	c := w.chunk(chunkPosFromBlockPos(pos))

	rid := BlockRuntimeID(b)

	var before uint32
	if rid != airRID && !opts.DisableLiquidDisplacement {
		before = c.Block(x, y, z, 0)
	}

	c.modified = true
	c.SetBlock(x, y, z, 0, rid)
	if nbtBlocks[rid] {
		c.BlockEntities[pos] = b
	} else {
		delete(c.BlockEntities, pos)
	}

	viewers := slices.Clone(c.viewers)

	if !opts.DisableLiquidDisplacement {
		var secondLayer Block

		if rid == airRID {
			if li := c.Block(x, y, z, 1); li != airRID {
				c.SetBlock(x, y, z, 0, li)
				c.SetBlock(x, y, z, 1, airRID)
				secondLayer = air()
				b = blockByRuntimeIDOrAir(li)
			}
		} else if liquidDisplacingBlocks[rid] && liquidBlocks[before] {
			l := blockByRuntimeIDOrAir(before)
			if b.(LiquidDisplacer).CanDisplace(l.(Liquid)) {
				c.SetBlock(x, y, z, 1, before)
				secondLayer = l
			}
		}

		if secondLayer != nil {
			for _, viewer := range viewers {
				viewer.ViewBlockUpdate(pos, secondLayer, 1)
			}
		}
	}

	for _, viewer := range viewers {
		viewer.ViewBlockUpdate(pos, b, 0)
	}

	if !opts.DisableBlockUpdates {
		w.doBlockUpdatesAround(pos)
	}
}

// SetBiome sets the biome at the position passed. If a chunk is not yet loaded at that position, the chunk is
// first loaded or generated if it could not be found in the world save.
func (w *World) setBiome(pos cube.Pos, b Biome) {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	c.modified = true
	c.SetBiome(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), uint32(b.EncodeBiome()))
}

// BuildStructure builds a Structure passed at a specific position in the world. Unlike SetBlock, it takes a
// Structure implementation, which provides blocks to be placed at a specific location.
// BuildStructure is specifically tinkered to be able to process a large batch of chunks simultaneously and
// will do so within much less time than separate SetBlock calls would.
// The method operates on a per-chunk basis, setting all blocks within a single chunk part of the structure
// before moving on to the next chunk.
func (w *World) buildStructure(pos cube.Pos, s Structure) {
	dim := s.Dimensions()
	width, height, length := dim[0], dim[1], dim[2]
	maxX, maxY, maxZ := pos[0]+width, pos[1]+height, pos[2]+length

	for chunkX := pos[0] >> 4; chunkX <= maxX>>4; chunkX++ {
		for chunkZ := pos[2] >> 4; chunkZ <= maxZ>>4; chunkZ++ {
			// We approach this on a per-chunk basis, so that we can keep only one chunk in memory at a time
			// while not needing to acquire a new chunk lock for every block. This also allows us not to send
			// block updates, but instead send a single chunk update once.
			chunkPos := ChunkPos{int32(chunkX), int32(chunkZ)}
			c := w.chunk(chunkPos)
			f := func(x, y, z int) Block {
				actual := cube.Pos{pos[0] + x, pos[1] + y, pos[2] + z}
				if actual[0]>>4 == chunkX && actual[2]>>4 == chunkZ {
					return w.blockInChunk(c, actual)
				}
				return w.block(actual)
			}
			baseX, baseZ := chunkX<<4, chunkZ<<4
			subs := c.Sub()
			for i, sub := range subs {
				baseY := (i + (w.Range()[0] >> 4)) << 4
				if baseY>>4 < pos[1]>>4 {
					continue
				} else if baseY >= maxY {
					break
				}

				for localY := 0; localY < 16; localY++ {
					yOffset := baseY + localY
					if yOffset > w.Range()[1] || yOffset >= maxY {
						// We've hit the height limit for blocks.
						break
					} else if yOffset < w.Range()[0] || yOffset < pos[1] {
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
								rid := BlockRuntimeID(b)
								sub.SetBlock(uint8(xOffset), uint8(yOffset), uint8(zOffset), 0, rid)

								nbtPos := cube.Pos{xOffset, yOffset, zOffset}
								if nbtBlocks[rid] {
									c.BlockEntities[nbtPos] = b
								} else {
									delete(c.BlockEntities, nbtPos)
								}
							}
							if liq != nil {
								sub.SetBlock(uint8(xOffset), uint8(yOffset), uint8(zOffset), 1, BlockRuntimeID(liq))
							} else if len(sub.Layers()) > 1 {
								sub.SetBlock(uint8(xOffset), uint8(yOffset), uint8(zOffset), 1, airRID)
							}
						}
					}
				}
			}
			c.SetBlock(0, 0, 0, 0, c.Block(0, 0, 0, 0)) // Make sure the heightmap is recalculated.
			c.modified = true

			// After setting all blocks of the structure within a single chunk,
			// we show the new chunk to all viewers once.
			for _, viewer := range c.viewers {
				viewer.ViewChunk(chunkPos, c.Chunk, c.BlockEntities)
			}
		}
	}
}

// Liquid attempts to return any liquid block at the position passed. This liquid may be in the foreground or
// in any other layer.
// If found, the liquid is returned. If not, the bool returned is false and the liquid is nil.
func (w *World) liquid(pos cube.Pos) (Liquid, bool) {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return nil, false
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])

	id := c.Block(x, y, z, 0)
	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.conf.Log.Error("Liquid: no block with runtime ID", "ID", id)
		return nil, false
	}
	if liq, ok := b.(Liquid); ok {
		return liq, true
	}
	id = c.Block(x, y, z, 1)

	b, ok = BlockByRuntimeID(id)
	if !ok {
		w.conf.Log.Error("Liquid: no block with runtime ID", "ID", id)
		return nil, false
	}
	liq, ok := b.(Liquid)
	return liq, ok
}

// SetLiquid sets the liquid at a specific position in the world. Unlike SetBlock, SetLiquid will not
// overwrite any existing blocks. It will instead be in the same position as a block currently there, unless
// there already is a liquid at that position, in which case it will be overwritten.
// If nil is passed for the liquid, any liquid currently present will be removed.
func (w *World) setLiquid(pos cube.Pos, b Liquid) {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return
	}
	chunkPos := chunkPosFromBlockPos(pos)
	c := w.chunk(chunkPos)
	if b == nil {
		w.removeLiquids(c, pos)
		w.doBlockUpdatesAround(pos)
		return
	}
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])
	if !replaceable(w, c, pos, b) {
		if displacer, ok := w.blockInChunk(c, pos).(LiquidDisplacer); !ok || !displacer.CanDisplace(b) {
			return
		}
	}
	rid := BlockRuntimeID(b)
	if w.removeLiquids(c, pos) {
		c.SetBlock(x, y, z, 0, rid)
		for _, v := range c.viewers {
			v.ViewBlockUpdate(pos, b, 0)
		}
	} else {
		c.SetBlock(x, y, z, 1, rid)
		for _, v := range c.viewers {
			v.ViewBlockUpdate(pos, b, 1)
		}
	}
	c.modified = true

	w.doBlockUpdatesAround(pos)
}

// removeLiquids removes any liquid blocks that may be present at a specific block position in the chunk
// passed.
// The bool returned specifies if no blocks were left on the foreground layer.
func (w *World) removeLiquids(c *Column, pos cube.Pos) bool {
	x, y, z := uint8(pos[0]), int16(pos[1]), uint8(pos[2])

	noneLeft := false
	if noLeft, changed := w.removeLiquidOnLayer(c.Chunk, x, y, z, 0); noLeft {
		if changed {
			for _, v := range c.viewers {
				v.ViewBlockUpdate(pos, air(), 0)
			}
		}
		noneLeft = true
	}
	if _, changed := w.removeLiquidOnLayer(c.Chunk, x, y, z, 1); changed {
		for _, v := range c.viewers {
			v.ViewBlockUpdate(pos, air(), 1)
		}
	}
	return noneLeft
}

// removeLiquidOnLayer removes a liquid block from a specific layer in the chunk passed, returning true if
// successful.
func (w *World) removeLiquidOnLayer(c *chunk.Chunk, x uint8, y int16, z, layer uint8) (bool, bool) {
	id := c.Block(x, y, z, layer)

	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.conf.Log.Error("removeLiquidOnLayer: no block with runtime ID", "ID", id)
		return false, false
	}
	if _, ok := b.(Liquid); ok {
		c.SetBlock(x, y, z, layer, airRID)
		return true, true
	}
	return id == airRID, false
}

// additionalLiquid checks if the block at a position has additional liquid on another layer and returns the
// liquid if so.
func (w *World) additionalLiquid(pos cube.Pos) (Liquid, bool) {
	if pos.OutOfBounds(w.Range()) {
		// Fast way out.
		return nil, false
	}
	c := w.chunk(chunkPosFromBlockPos(pos))
	id := c.Block(uint8(pos[0]), int16(pos[1]), uint8(pos[2]), 1)

	b, ok := BlockByRuntimeID(id)
	if !ok {
		w.conf.Log.Error("additionalLiquid: no block with runtime ID", "ID", id)
		return nil, false
	}
	liq, ok := b.(Liquid)
	return liq, ok
}

// Light returns the light level at the position passed. This is the highest of the sky and block light.
// The light value returned is a value in the range 0-15, where 0 means there is no light present, whereas
// 15 means the block is fully lit.
func (w *World) light(pos cube.Pos) uint8 {
	if pos[1] < w.ra[0] {
		// Fast way out.
		return 0
	}
	if pos[1] > w.ra[1] {
		// Above the rest of the world, so full skylight.
		return 15
	}
	return w.chunk(chunkPosFromBlockPos(pos)).Light(uint8(pos[0]), int16(pos[1]), uint8(pos[2]))
}

// SkyLight returns the skylight level at the position passed. This light level
// is not influenced by blocks that emit light, such as torches. The light
// value, similarly to Light, is a value in the range 0-15, where 0 means no
// light is present.
func (w *World) skyLight(pos cube.Pos) uint8 {
	if pos[1] < w.ra[0] {
		// Fast way out.
		return 0
	}
	if pos[1] > w.ra[1] {
		// Above the rest of the world, so full skylight.
		return 15
	}
	return w.chunk(chunkPosFromBlockPos(pos)).SkyLight(uint8(pos[0]), int16(pos[1]), uint8(pos[2]))
}

// Time returns the current time of the world. The time is incremented every 1/20th of a second, unless
// World.StopTime() is called.
func (w *World) Time() int {
	if w == nil {
		return 0
	}
	w.set.Lock()
	defer w.set.Unlock()
	return int(w.set.Time)
}

// SetTime sets the new time of the world. SetTime will always work, regardless of whether the time is stopped
// or not.
func (w *World) SetTime(new int) {
	if w == nil {
		return
	}
	w.set.Lock()
	w.set.Time = int64(new)
	w.set.Unlock()

	viewers, _ := w.allViewers()
	for _, viewer := range viewers {
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
	w.set.Lock()
	defer w.set.Unlock()
	w.set.TimeCycle = v
}

// Temperature returns the temperature in the World at a specific position. Higher altitudes and different biomes
// influence the temperature returned.
func (w *World) temperature(pos cube.Pos) float64 {
	const (
		tempDrop = 1.0 / 600
		seaLevel = 64
	)
	diff := max(pos[1]-seaLevel, 0)
	return w.biome(pos).Temperature() - float64(diff)*tempDrop
}

// AddParticle spawns a particle at a given position in the world. Viewers that are viewing the chunk will be
// shown the particle.
func (w *World) addParticle(pos mgl64.Vec3, p Particle) {
	p.Spawn(w, pos)
	for _, viewer := range w.viewersOf(pos) {
		viewer.ViewParticle(pos, p)
	}
}

// PlaySound plays a sound at a specific position in the world. Viewers of that position will be able to hear
// the sound if they're close enough.
func (w *World) playSound(tx *Tx, pos mgl64.Vec3, s Sound) {
	ctx := event.C(tx)
	if w.Handler().HandleSound(ctx, s, pos); ctx.Cancelled() {
		return
	}
	for _, viewer := range w.viewersOf(pos) {
		viewer.ViewSound(pos, s)
	}
}

// AddEntity adds an Entity to the world at the position that the Entity has. The Entity will be visible to
// all viewers of the world that have the chunk of the Entity loaded.
// If the chunk that the Entity is in is not yet loaded, it will first be loaded.
// If the Entity passed to AddEntity is currently in a world, it is first removed from that world.
func (w *World) addEntity(tx *Tx, handle *EntityHandle) Entity {
	handle.setAndUnlockWorld(w, tx)
	pos := chunkPosFromVec3(handle.data.Pos)
	w.entities[handle] = pos

	c := w.chunk(pos)
	c.Entities, c.modified = append(c.Entities, handle), true

	e := handle.mustEntity(tx)
	for _, v := range c.viewers {
		// We show the entity to all viewers currently in the chunk that the entity is spawned in.
		showEntity(e, v)
	}
	w.Handler().HandleEntitySpawn(e)
	return e
}

// RemoveEntity removes an Entity from the world that is currently present in it. Any viewers of the Entity
// will no longer be able to see it.
// RemoveEntity operates assuming the position of the Entity is the same as where it is currently in the
// world. If it can not find it there, it will loop through all entities and try to find it.
// RemoveEntity assumes the Entity is currently loaded and in a loaded chunk. If not, the function will not do
// anything.
func (w *World) removeEntity(e Entity, tx *Tx) *EntityHandle {
	handle := e.H()
	pos, found := w.entities[handle]
	if !found {
		// The entity currently isn't in this world.
		return nil
	}
	w.Handler().HandleEntityDespawn(e)

	c := w.chunk(pos)
	c.Entities, c.modified = sliceutil.DeleteVal(c.Entities, handle), true

	for _, v := range c.viewers {
		v.HideEntity(e)
	}
	delete(w.entities, handle)
	handle.unsetAndLockWorld(tx)
	return handle
}

// EntitiesWithin does a lookup through the entities in the chunks touched by the BBox passed, returning all
// those which are contained within the BBox when it comes to their position.
func (w *World) entitiesWithin(tx *Tx, box cube.BBox) iter.Seq[Entity] {
	return func(yield func(Entity) bool) {
		minPos, maxPos := chunkPosFromVec3(box.Min()), chunkPosFromVec3(box.Max())

		for x := minPos[0]; x <= maxPos[0]; x++ {
			for z := minPos[1]; z <= maxPos[1]; z++ {
				c, ok := w.chunks[ChunkPos{x, z}]
				if !ok {
					// The chunk wasn't loaded, so there are no entities here.
					continue
				}
				for _, handle := range c.Entities {
					if !box.Vec3Within(handle.data.Pos) {
						continue
					}
					if !yield(handle.mustEntity(tx)) {
						return
					}
				}
			}
		}
	}
}

// allEntities returns an iterator that yields all entities in the World.
func (w *World) allEntities(tx *Tx) iter.Seq[Entity] {
	return func(yield func(Entity) bool) {
		for e := range w.entities {
			if ent := e.mustEntity(tx); !yield(ent) {
				return
			}
		}
	}
}

// allPlayers returns an iterator that yields all player entities in the World.
func (w *World) allPlayers(tx *Tx) iter.Seq[Entity] {
	return func(yield func(Entity) bool) {
		for e := range w.entities {
			if [8]byte(e.id[:8]) != [8]byte(uuid.Nil[:8]) {
				// The first 8 bytes are set, so we are dealing with a player.
				// Non-player entities only use the last 8 bytes.
				if ent := e.mustEntity(tx); !yield(ent) {
					return
				}
			}
		}
	}
}

// Spawn returns the spawn of the world. Every new player will by default spawn on this position in the world
// when joining.
func (w *World) Spawn() cube.Pos {
	if w == nil {
		return cube.Pos{}
	}
	w.set.Lock()
	s := w.set.Spawn
	w.set.Unlock()
	if s[1] > w.Range()[1] {
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
	w.set.Lock()
	w.set.Spawn = pos
	w.set.Unlock()

	viewers, _ := w.allViewers()
	for _, viewer := range viewers {
		viewer.ViewWorldSpawn(pos)
	}
}

// PlayerSpawn returns the spawn position of a player with a UUID in this World.
func (w *World) PlayerSpawn(id uuid.UUID) cube.Pos {
	if w == nil {
		return cube.Pos{}
	}
	pos, exist, err := w.conf.Provider.LoadPlayerSpawnPosition(id)
	if err != nil {
		w.conf.Log.Error("load player spawn: "+err.Error(), "ID", id)
		return w.Spawn()
	}
	if !exist {
		return w.Spawn()
	}
	return pos
}

// SetPlayerSpawn sets the spawn position of a player with a UUID in this World. If the player has a spawn in the world,
// the player will be teleported to this location on respawn.
func (w *World) SetPlayerSpawn(id uuid.UUID, pos cube.Pos) {
	if w == nil {
		return
	}
	if err := w.conf.Provider.SavePlayerSpawnPosition(id, pos); err != nil {
		w.conf.Log.Error("save player spawn: "+err.Error(), "ID", id)
	}
}

// DefaultGameMode returns the default game mode of the world. When players join, they are given this game
// mode.
// The default game mode may be changed using SetDefaultGameMode().
func (w *World) DefaultGameMode() GameMode {
	if w == nil {
		return GameModeSurvival
	}
	w.set.Lock()
	defer w.set.Unlock()
	return w.set.DefaultGameMode
}

// SetTickRange sets the range in chunks around each Viewer that will have the chunks (their blocks and entities)
// ticked when the World is ticked.
func (w *World) SetTickRange(v int) {
	if w == nil {
		return
	}
	w.set.Lock()
	defer w.set.Unlock()
	w.set.TickRange = int32(v)
}

// tickRange returns the tick range around each Viewer.
func (w *World) tickRange() int {
	w.set.Lock()
	defer w.set.Unlock()
	return int(w.set.TickRange)
}

// SetDefaultGameMode changes the default game mode of the world. When players join, they are then given that
// game mode.
func (w *World) SetDefaultGameMode(mode GameMode) {
	if w == nil {
		return
	}
	w.set.Lock()
	defer w.set.Unlock()
	w.set.DefaultGameMode = mode
}

// Difficulty returns the difficulty of the world. Properties of mobs in the world and the player's hunger
// will depend on this difficulty.
func (w *World) Difficulty() Difficulty {
	if w == nil {
		return DifficultyNormal
	}
	w.set.Lock()
	defer w.set.Unlock()
	return w.set.Difficulty
}

// SetDifficulty changes the difficulty of a world.
func (w *World) SetDifficulty(d Difficulty) {
	if w == nil {
		return
	}
	w.set.Lock()
	defer w.set.Unlock()
	w.set.Difficulty = d
}

// ScheduleBlockUpdate schedules a block update at the position passed after a specific delay. If the block at
// that position does not handle block updates, nothing will happen.
func (w *World) scheduleBlockUpdate(pos cube.Pos, delay time.Duration) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		return
	}
	if _, exists := w.scheduledUpdates[pos]; exists {
		return
	}
	w.set.Lock()
	t := w.set.CurrentTick
	w.set.Unlock()

	w.scheduledUpdates[pos] = t + delay.Nanoseconds()/int64(time.Second/20)
}

// doBlockUpdatesAround schedules block updates directly around and on the position passed.
func (w *World) doBlockUpdatesAround(pos cube.Pos) {
	if w == nil || pos.OutOfBounds(w.Range()) {
		return
	}
	changed := pos

	w.updateNeighbour(pos, changed)
	pos.Neighbours(func(pos cube.Pos) {
		w.updateNeighbour(pos, changed)
	}, w.Range())
}

// neighbourUpdate represents a position that needs to be updated because of a neighbour that changed.
type neighbourUpdate struct {
	pos, neighbour cube.Pos
}

// updateNeighbour ticks the position passed as a result of the neighbour passed being updated.
func (w *World) updateNeighbour(pos, changedNeighbour cube.Pos) {
	w.neighbourUpdates = append(w.neighbourUpdates, neighbourUpdate{pos: pos, neighbour: changedNeighbour})
}

// Handle changes the current Handler of the world. As a result, events called by the world will call
// handlers of the Handler passed.
// Handle sets the world's Handler to NopHandler if nil is passed.
func (w *World) Handle(h Handler) {
	if w == nil {
		return
	}
	if h == nil {
		h = NopHandler{}
	}
	w.handler.Store(&h)
}

// Viewers returns a list of all viewers viewing the position passed. A viewer will be assumed to be watching
// if the position is within one of the chunks that the viewer is watching.
func (w *World) viewersOf(pos mgl64.Vec3) (viewers []Viewer) {
	c, ok := w.chunks[chunkPosFromVec3(pos)]
	if !ok {
		return nil
	}
	return c.viewers
}

// PortalDestination returns the destination world for a portal of a specific Dimension. If no destination World could
// be found, the current World is returned.
func (w *World) PortalDestination(dim Dimension) *World {
	if w.conf.PortalDestination == nil {
		return w
	}
	if res := w.conf.PortalDestination(dim); res != nil {
		return res
	}
	return w
}

// Save saves the World to the provider.
func (w *World) Save() {
	<-w.Exec(w.save)
}

func (w *World) save(tx *Tx) {
	w.conf.Log.Debug("Saving chunks in memory to disk...")
	for pos, c := range w.chunks {
		w.saveChunk(tx, pos, c, false)
	}
}

// Close closes the world and saves all chunks currently loaded.
func (w *World) Close() error {
	w.o.Do(w.close)
	return nil
}

// close stops the World from ticking, saves all chunks to the Provider and updates the world's settings.
func (w *World) close() {
	// Let user code run anything that needs to be finished before the World is closed.
	w.Handler().HandleClose()
	w.Handle(NopHandler{})

	w.Save()
	maps.Clear(w.chunks)

	close(w.closing)
	w.running.Wait()

	w.set.ref.Add(-1)
	if !w.advance {
		return
	}

	if !w.conf.ReadOnly {
		w.conf.Log.Debug("Updating level.dat values...")

		w.provider().SaveSettings(w.set)
	}

	w.conf.Log.Debug("Closing provider...")
	if err := w.provider().Close(); err != nil {
		w.conf.Log.Error("close world provider: " + err.Error())
	}
}

// allViewers returns a list of all loaders of the world, regardless of where in the world they are viewing.
func (w *World) allViewers() ([]Viewer, []*Loader) {
	viewers, loaders := make([]Viewer, 0, len(w.viewers)), make([]*Loader, 0, len(w.viewers))
	for k, v := range w.viewers {
		viewers = append(viewers, v)
		loaders = append(loaders, k)
	}
	return viewers, loaders
}

// addWorldViewer adds a viewer to the world. Should only be used while the viewer isn't viewing any chunks.
func (w *World) addWorldViewer(l *Loader) {
	w.viewers[l] = l.viewer
	l.viewer.ViewTime(w.Time())
	w.set.Lock()
	raining, thundering := w.set.Raining, w.set.Raining && w.set.Thundering
	w.set.Unlock()
	l.viewer.ViewWeather(raining, thundering)
	l.viewer.ViewWorldSpawn(w.Spawn())
}

// addViewer adds a viewer to the world at a given position. Any events that happen in the chunk at that
// position, such as block changes, Entity changes etc., will be sent to the viewer.
func (w *World) addViewer(tx *Tx, c *Column, loader *Loader) {
	c.viewers = append(c.viewers, loader.viewer)
	c.loaders = append(c.loaders, loader)

	for _, entity := range c.Entities {
		showEntity(entity.mustEntity(tx), loader.viewer)
	}
}

// removeViewer removes a viewer from the world at a given position. All entities will be hidden from the
// viewer and no more calls will be made when events in the chunk happen.
func (w *World) removeViewer(tx *Tx, pos ChunkPos, loader *Loader) {
	if w == nil {
		return
	}
	c, ok := w.chunks[pos]
	if !ok {
		return
	}
	if i := slices.Index(c.loaders, loader); i != -1 {
		c.viewers = slices.Delete(c.viewers, i, i+1)
		c.loaders = slices.Delete(c.loaders, i, i+1)
	}

	// After removing the loader from the chunk, we also need to hide all entities from the viewer.
	for _, entity := range c.Entities {
		loader.viewer.HideEntity(entity.mustEntity(tx))
	}
}

// provider returns the provider of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) provider() Provider {
	return w.conf.Provider
}

// Handler returns the Handler of the world. It should always be used, rather than direct field access, in
// order to provide synchronisation safety.
func (w *World) Handler() Handler {
	if w == nil {
		return NopHandler{}
	}
	return *w.handler.Load()
}

// showEntity shows an Entity to a viewer of the world. It makes sure everything of the Entity, including the
// items held, is shown.
func showEntity(e Entity, viewer Viewer) {
	viewer.ViewEntity(e)
	viewer.ViewEntityItems(e)
	viewer.ViewEntityArmour(e)
}

// chunk reads a chunk from the position passed. If a chunk at that position is not yet loaded, the chunk is
// loaded from the provider, or generated if it did not yet exist. Both of these actions are done
// synchronously.
// An error is returned if the chunk could not be loaded successfully.
// chunk locks the chunk returned, meaning that any call to chunk made at the same time has to wait until the
// user calls Chunk.Unlock() on the chunk returned.
func (w *World) chunk(pos ChunkPos) *Column {
	c, ok := w.chunks[pos]
	if ok {
		return c
	}
	c, err := w.loadChunk(pos)
	chunk.LightArea([]*chunk.Chunk{c.Chunk}, int(pos[0]), int(pos[1])).Fill()
	if err != nil {
		w.conf.Log.Error("load chunk: "+err.Error(), "X", pos[0], "Z", pos[1])
		return c
	}
	w.calculateLight(pos)
	return c
}

// loadChunk attempts to load a chunk from the provider, or generates a chunk if one doesn't currently exist.
func (w *World) loadChunk(pos ChunkPos) (*Column, error) {
	column, err := w.provider().LoadColumn(pos, w.conf.Dim)
	col := columnFrom(column, w)
	switch {
	case err == nil:
		w.chunks[pos] = col
		for _, e := range col.Entities {
			w.entities[e] = pos
			e.w = w
		}
		return col, nil
	case errors.Is(err, leveldb.ErrNotFound):
		// The provider doesn't have a chunk saved at this position, so we generate a new one.
		col = newColumn(chunk.New(airRID, w.Range()))
		w.chunks[pos] = col

		w.conf.Generator.GenerateChunk(pos, col.Chunk)
		return col, nil
	default:
		return newColumn(chunk.New(airRID, w.Range())), err
	}
}

// calculateLight calculates the light in the chunk passed and spreads the
// light of any surrounding neighbours if they have all chunks loaded around it
// as a result of the one passed.
func (w *World) calculateLight(centre ChunkPos) {
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			// For all the neighbours of this chunk, if they exist, check if all
			// neighbours of that chunk now exist because of this one.
			pos := ChunkPos{centre[0] + x, centre[1] + z}
			if _, ok := w.chunks[pos]; ok {
				// Attempt to spread the light of all neighbours into the
				// surrounding ones.
				w.spreadLight(pos)
			}
		}
	}
}

// spreadLight spreads the light from the chunk passed at the position passed to all neighbours if each of
// them is loaded.
func (w *World) spreadLight(pos ChunkPos) {
	c := make([]*chunk.Chunk, 0, 9)
	for z := int32(-1); z <= 1; z++ {
		for x := int32(-1); x <= 1; x++ {
			neighbour, ok := w.chunks[ChunkPos{pos[0] + x, pos[1] + z}]
			if !ok {
				// Not all surrounding chunks existed: Stop spreading light as we can't do it completely yet.
				return
			}
			c = append(c, neighbour.Chunk)
		}
	}
	// All chunks of the current one are present, so we can spread the light from this chunk
	// to all chunks.
	chunk.LightArea(c, int(pos[0])-1, int(pos[1])-1).Spread()
}

// saveChunk is called when a chunk is removed from the cache. We first compact the chunk, then we write it to
// the provider.
func (w *World) saveChunk(tx *Tx, pos ChunkPos, c *Column, closeEntities bool) {
	if !w.conf.ReadOnly && c.modified {
		c.Compact()
		if err := w.provider().StoreColumn(pos, w.conf.Dim, columnTo(c, tx)); err != nil {
			w.conf.Log.Error("save chunk: "+err.Error(), "X", pos[0], "Z", pos[1])
		}
	}
	if closeEntities {
		for _, e := range c.Entities {
			_ = e.mustEntity(tx).Close()
		}
		clear(c.Entities)
	}
}

// chunkCacheJanitor runs until the world is running, cleaning chunks that are no longer in use from the cache.
func (w *World) chunkCacheJanitor() {
	t := time.NewTicker(time.Minute * 5)
	defer t.Stop()

	w.running.Add(1)
	for {
		select {
		case <-t.C:
			<-w.Exec(w.closeUnusedChunks)
		case <-w.closing:
			w.running.Done()
			return
		}
	}
}

func (w *World) closeUnusedChunks(tx *Tx) {
	for pos, c := range w.chunks {
		if len(c.viewers) == 0 {
			delete(w.chunks, pos)
			w.saveChunk(tx, pos, c, true)
		}
	}
}

// Column represents the data of a chunk including the block entities and loaders. This data is protected
// by the mutex present in the chunk.Chunk held.
type Column struct {
	modified bool

	*chunk.Chunk
	Entities      []*EntityHandle
	BlockEntities map[cube.Pos]Block

	viewers []Viewer
	loaders []*Loader
}

// newColumn returns a new Column wrapper around the chunk.Chunk passed.
func newColumn(c *chunk.Chunk) *Column {
	return &Column{Chunk: c, BlockEntities: map[cube.Pos]Block{}}
}

func columnTo(col *Column, tx *Tx) *chunk.Column {
	c := &chunk.Column{
		Chunk:         col.Chunk,
		Entities:      make([]chunk.Entity, 0, len(col.Entities)),
		BlockEntities: make([]chunk.BlockEntity, 0, len(col.BlockEntities)),
	}
	for _, e := range col.Entities {
		data := e.encodeNBT()
		maps.Copy(data, e.t.EncodeNBT(&e.data))
		data["identifier"] = e.t.EncodeEntity()
		c.Entities = append(c.Entities, chunk.Entity{ID: int64(binary.LittleEndian.Uint64(e.id[8:])), Data: data})
	}
	for pos, be := range col.BlockEntities {
		c.BlockEntities = append(c.BlockEntities, chunk.BlockEntity{Pos: pos, Data: be.(NBTer).EncodeNBT()})
	}
	return c
}

func columnFrom(c *chunk.Column, w *World) *Column {
	col := &Column{
		Chunk:         c.Chunk,
		Entities:      make([]*EntityHandle, 0, len(c.Entities)),
		BlockEntities: make(map[cube.Pos]Block, len(c.BlockEntities)),
	}
	for _, e := range c.Entities {
		eid, ok := e.Data["identifier"].(string)
		if !ok {
			w.conf.Log.Error("read column: entity without identifier field", "ID", e.ID)
			continue
		}
		t, ok := w.conf.Entities.Lookup(eid)
		if !ok {
			w.conf.Log.Error("read column: unknown entity type", "ID", e.ID, "type", eid)
			continue
		}
		col.Entities = append(col.Entities, entityFromData(t, e.ID, e.Data))
	}
	for _, be := range c.BlockEntities {
		rid := c.Chunk.Block(uint8(be.Pos[0]), int16(be.Pos[1]), uint8(be.Pos[2]), 0)
		b, ok := BlockByRuntimeID(rid)
		if !ok {
			w.conf.Log.Error("read column: no block with runtime ID", "ID", rid)
			continue
		}
		nb, ok := b.(NBTer)
		if !ok {
			w.conf.Log.Error("read column: block with nbt does not implement NBTer", "block", fmt.Sprintf("%#v", b))
			continue
		}
		col.BlockEntities[be.Pos] = nb.DecodeNBT(be.Data).(Block)
	}

	return col
}
