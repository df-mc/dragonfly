package world

import (
	"maps"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
)

// ticker implements World ticking methods.
type ticker struct {
	interval time.Duration
}

// tickLoop starts ticking the World 20 times every second, updating all
// entities, blocks and other features such as the time and weather of the
// world, as required.
func (t ticker) tickLoop(w *World) {
	tc := time.NewTicker(t.interval)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			<-w.exec(t.tick)
		case <-w.closing:
			// World is being closed: Stop ticking and get rid of a task.
			w.running.Done()
			return
		}
	}
}

// AdvanceTick advances the World by a single tick. It is generally only useful
// for Worlds created with Config.Synchronous set: other Worlds tick
// automatically 20 times per second. Synchronous Worlds tick loaded chunks
// even when no viewers are present.
func (w *World) AdvanceTick() {
	<-w.exec(ticker{}.tick)
}

// tick performs a tick on the World and updates the time, weather, blocks and
// entities that require updates.
func (t ticker) tick(tx *Tx) {
	viewers, loaders := tx.World().allViewers()
	w := tx.World()

	w.set.Lock()
	if s := w.set.Spawn; s[1] > tx.Range()[1] && w.Dimension() == Overworld {
		// Vanilla will set the spawn position's Y value to max to indicate that
		// the player should spawn at the highest position in the world.
		w.set.Spawn[1] = w.highestObstructingBlock(s[0], s[2]) + 1
	}
	if len(viewers) == 0 && w.set.CurrentTick != 0 && !w.conf.Synchronous {
		// Don't continue ticking if no viewers are in the world. Synchronous
		// worlds only tick on explicit AdvanceTick calls, so they always tick.
		w.set.Unlock()
		return
	}
	if w.advance {
		w.set.CurrentTick++
		if w.set.TimeCycle {
			w.set.Time++
		}
		if w.set.WeatherCycle {
			w.advanceWeather()
		}
	}

	rain, thunder, tick, tim, cycle := w.set.Raining, w.set.Thundering && w.set.Raining, w.set.CurrentTick, int(w.set.Time), w.set.TimeCycle

	tryAdvanceDay := false
	if tx.w.set.RequiredSleepTicks > 0 {
		tx.w.set.RequiredSleepTicks--
		tryAdvanceDay = tx.w.set.RequiredSleepTicks <= 0
	}

	w.set.Unlock()

	if tryAdvanceDay {
		t.tryAdvanceDay(tx, cycle)
	}

	if tick%20 == 0 {
		for _, viewer := range viewers {
			if w.Dimension().TimeCycle() && cycle {
				viewer.ViewTime(tim)
			}
			if w.Dimension().WeatherCycle() {
				viewer.ViewWeather(rain, thunder)
			}
		}
	}
	if thunder {
		w.tickLightning(tx)
	}

	t.tickEntities(tx, tick)
	w.scheduledUpdates.tick(tx, tick)
	t.tickBlocksRandomly(tx, loaders, tick)
	t.performNeighbourUpdates(tx)
	w.redstone.tick(tx, tick)
}

// performNeighbourUpdates performs all block updates that came as a result of a neighbouring block being changed.
func (t ticker) performNeighbourUpdates(tx *Tx) {
	updates := slices.Clone(tx.World().neighbourUpdates)
	clear(tx.World().neighbourUpdates)
	tx.World().neighbourUpdates = tx.World().neighbourUpdates[:0]

	for _, update := range updates {
		pos, changedNeighbour := update.pos, update.neighbour
		if ticker, ok := tx.Block(pos).(NeighbourUpdateTicker); ok {
			ticker.NeighbourUpdateTick(pos, changedNeighbour, tx)
		}
		if liquid, ok := tx.World().additionalLiquid(pos); ok {
			if ticker, ok := liquid.(NeighbourUpdateTicker); ok {
				ticker.NeighbourUpdateTick(pos, changedNeighbour, tx)
			}
		}
	}
}

// tickBlocksRandomly executes random block ticks in loaded chunks within range of loaders. The read-only
// chunk scan is gathered in parallel per tick region; all block callbacks run on the owner afterwards.
func (t ticker) tickBlocksRandomly(tx *Tx, loaders []*Loader, tick int64) {
	w := tx.World()
	r := int32(w.tickRange())
	if r == 0 {
		// NOP if the simulation distance is 0.
		return
	}

	// Synchronous worlds tick all loaded chunks, regardless of loaders.
	all := w.conf.Synchronous
	var loaded []ChunkPos
	if !all {
		loaded = make([]ChunkPos, 0, len(loaders))
		for _, loader := range loaders {
			loader.mu.RLock()
			pos := loader.pos
			loader.mu.RUnlock()

			loaded = append(loaded, pos)
		}
		if len(loaded) == 0 {
			// Without loaders, no chunk can be within simulation distance.
			return
		}
	}

	n := 0
	for _, reg := range w.tickRegions.all(w.chunks) {
		state := regionIn
		if !all {
			if state = reg.rangeState(loaded, r); state == regionOut {
				// The whole region is out of simulation distance of every loader.
				continue
			}
		}
		for chunks := range slices.Chunk(reg.chunks, regionBatchSize) {
			if n == len(w.tickBatches) {
				w.tickBatches = append(w.tickBatches, tickBatch{})
			}
			b := &w.tickBatches[n]
			n++
			b.chunks, b.checkRange = chunks, state == regionPartial
			b.blockEntities, b.randomBlocks = b.blockEntities[:0], b.randomBlocks[:0]
			// Seeding batches serially up front keeps results deterministic regardless of scheduling.
			b.pcg.Seed(w.r.Uint64(), w.r.Uint64())
		}
	}
	// Drop chunk references held by scratch batches unused this tick.
	for i := n; i < len(w.tickBatches); i++ {
		w.tickBatches[i].chunks = nil
	}
	batches := w.tickBatches[:n]

	runTickJobs(w.conf.TickWorkers, batches, func(b *tickBatch) {
		t.gatherTickCandidates(w, b, loaded, r)
	})

	for i := range batches {
		for _, pos := range batches[i].randomBlocks {
			if rb, ok := tx.Block(pos).(RandomTicker); ok {
				rb.RandomTick(pos, tx, w.r)
			}
		}
	}
	for i := range batches {
		for _, pos := range batches[i].blockEntities {
			if tb, ok := tx.Block(pos).(TickerBlock); ok {
				tb.Tick(tick, pos, tx)
			}
		}
	}
}

// gatherTickCandidates scans a batch's chunks for random tick candidates and ticking block entities. It
// runs concurrently with other batches and must only read World state, never mutate it.
func (t ticker) gatherTickCandidates(w *World, b *tickBatch, loaded []ChunkPos, r int32) {
	var g randUint4
	for _, rc := range b.chunks {
		if b.checkRange && !t.anyWithinDistance(rc.pos, loaded, r) {
			// No loaders in this chunk that are within the simulation distance, so proceed to the next.
			continue
		}
		b.blockEntities = slices.AppendSeq(b.blockEntities, maps.Keys(rc.col.BlockEntities))

		cx, cz := int(rc.pos[0]<<4), int(rc.pos[1]<<4)

		// We generate up to j random positions for every sub chunk.
		for j := 0; j < w.conf.RandomTickSpeed; j++ {
			x, y, z := g.uint4(&b.pcg), g.uint4(&b.pcg), g.uint4(&b.pcg)

			for i, sub := range rc.col.Sub() {
				if sub.Empty() {
					// SubChunk is empty, so skip it right away.
					continue
				}
				// Generally we would want to make sure the block has its block entities, but provided blocks
				// with block entities are generally ticked already, we are safe to assume that blocks
				// implementing the RandomTicker don't rely on additional block entity data.
				if rid := sub.Layers()[0].At(x, y, z); w.conf.Blocks.RandomTickBlock(rid) {
					subY := (i + (w.ra.Min() >> 4)) << 4
					b.randomBlocks = append(b.randomBlocks, cube.Pos{cx + int(x), subY + int(y), cz + int(z)})

					// Only generate new coordinates if a tickable block was actually found. If not, we can just re-use
					// the coordinates for the next sub chunk.
					x, y, z = g.uint4(&b.pcg), g.uint4(&b.pcg), g.uint4(&b.pcg)
				}
			}
		}
	}
}

// anyWithinDistance checks if any of the ChunkPos loaded are within the distance r of the ChunkPos pos.
func (t ticker) anyWithinDistance(pos ChunkPos, loaded []ChunkPos, r int32) bool {
	for _, chunkPos := range loaded {
		xDiff, zDiff := chunkPos[0]-pos[0], chunkPos[1]-pos[1]
		if (xDiff*xDiff)+(zDiff*zDiff) <= r*r {
			// The chunk was within the simulation distance of at least one viewer, so we can proceed to
			// ticking the block.
			return true
		}
	}
	return false
}

// tickEntities ticks all entities in the world, making sure they are still located in the correct chunks and
// updating where necessary.
func (t ticker) tickEntities(tx *Tx, tick int64) {
	for handle, lastPos := range tx.World().entities {
		e := handle.mustEntity(tx)
		chunkPos := chunkPosFromVec3(handle.data.Pos)

		c, ok := tx.World().chunks[chunkPos]
		if !ok {
			continue
		}

		if lastPos != chunkPos {
			// The entity was stored using an outdated chunk position. We update it and make sure it is ready
			// for loaders to view it.
			tx.World().entities[handle] = chunkPos
			c.Entities = append(c.Entities, handle)

			var viewers []Viewer

			// When changing an entity's world, then teleporting it immediately, we could end up in a situation
			// where the old chunk of the entity was not loaded. In this case, it should be safe simply to ignore
			// the loaders from the old chunk. We can assume they never saw the entity in the first place.
			if old, ok := tx.World().chunks[lastPos]; ok {
				old.Entities = sliceutil.DeleteVal(old.Entities, handle)
				viewers = old.viewers
			}

			for _, viewer := range viewers {
				if slices.Index(c.viewers, viewer) == -1 {
					// First we hide the entity from all loaders that were previously viewing it, but no
					// longer are.
					viewer.HideEntity(e)
				}
			}
			for _, viewer := range c.viewers {
				if slices.Index(viewers, viewer) == -1 {
					// Then we show the entity to all loaders that are now viewing the entity in the new
					// chunk.
					showEntity(e, viewer)
				}
			}
		}

		if tx.World().conf.Synchronous || len(c.viewers) > 0 {
			if te, ok := e.(TickerEntity); ok {
				te.Tick(tx, tick)
			}
		}
	}
}

// randUint4 is a structure used to generate random uint4s.
type randUint4 struct {
	x uint64
	n uint8
}

// uint4 returns a random uint4.
func (g *randUint4) uint4(src rand.Source) uint8 {
	if g.n == 0 {
		g.x = src.Uint64()
		g.n = 16
	}
	val := g.x & 0b1111

	g.x >>= 4
	g.n--
	return uint8(val)
}

// scheduledTickQueue implements a queue for scheduled block updates. Scheduled
// block updates are both position and block type specific.
type scheduledTickQueue struct {
	ticks         []scheduledTick
	furthestTicks map[scheduledTickIndex]int64
	currentTick   int64
}

type scheduledTick struct {
	pos   cube.Pos
	b     Block
	bhash uint64
	t     int64
}

type scheduledTickIndex struct {
	pos  cube.Pos
	hash uint64
}

// newScheduledTickQueue creates a queue for scheduled block ticks.
func newScheduledTickQueue(tick int64) *scheduledTickQueue {
	return &scheduledTickQueue{furthestTicks: make(map[scheduledTickIndex]int64), currentTick: tick}
}

// tick processes scheduled ticks, calling ScheduledTicker.ScheduledTick for any
// block update that is scheduled for the tick passed, and removing it from the
// queue.
func (queue *scheduledTickQueue) tick(tx *Tx, tick int64) {
	queue.currentTick = tick

	w := tx.World()
	for _, t := range queue.ticks {
		if t.t > tick {
			continue
		}
		b := tx.Block(t.pos)
		if ticker, ok := b.(ScheduledTicker); ok && w.conf.Blocks.BlockHash(b) == t.bhash {
			ticker.ScheduledTick(t.pos, tx, w.r)
		} else if liquid, ok := tx.World().additionalLiquid(t.pos); ok && w.conf.Blocks.BlockHash(liquid) == t.bhash {
			if ticker, ok := liquid.(ScheduledTicker); ok {
				ticker.ScheduledTick(t.pos, tx, w.r)
			}
		}
	}

	// Clear scheduled ticks that were processed from the queue.
	queue.ticks = slices.DeleteFunc(queue.ticks, func(t scheduledTick) bool {
		return t.t <= tick
	})
	maps.DeleteFunc(queue.furthestTicks, func(index scheduledTickIndex, t int64) bool {
		return t <= tick
	})
}

// schedule schedules a block update at the position passed for the block type
// passed after a specific delay. A block update is only scheduled if no block
// update with the same position and block type is already scheduled at a later
// time than the newly scheduled update.
func (queue *scheduledTickQueue) schedule(br BlockRegistry, pos cube.Pos, b Block, delay time.Duration) {
	resTick := queue.currentTick + int64(max(delay/(time.Second/20), 1))
	index := scheduledTickIndex{pos: pos, hash: br.BlockHash(b)}
	if t, ok := queue.furthestTicks[index]; ok && t >= resTick && t > queue.currentTick {
		return
	}
	queue.furthestTicks[index] = resTick
	queue.ticks = append(queue.ticks, scheduledTick{pos: pos, t: resTick, b: b, bhash: index.hash})
}

// fromChunk returns all scheduled ticks positioned within a ChunkPos.
func (queue *scheduledTickQueue) fromChunk(pos ChunkPos) []scheduledTick {
	m := make([]scheduledTick, 0, 8)
	for _, t := range queue.ticks {
		if pos == chunkPosFromBlockPos(t.pos) {
			m = append(m, t)
		}
	}
	return m
}

// removeChunk removes all scheduled ticks positioned within a ChunkPos.
func (queue *scheduledTickQueue) removeChunk(pos ChunkPos) {
	queue.ticks = slices.DeleteFunc(queue.ticks, func(tick scheduledTick) bool {
		return chunkPosFromBlockPos(tick.pos) == pos
	})
	maps.DeleteFunc(queue.furthestTicks, func(index scheduledTickIndex, _ int64) bool {
		return chunkPosFromBlockPos(index.pos) == pos
	})
}

// add adds a slice of scheduled ticks to the queue. It assumes no duplicate
// ticks are present in the slice.
func (queue *scheduledTickQueue) add(ticks []scheduledTick) {
	queue.ticks = append(queue.ticks, ticks...)
	for _, t := range ticks {
		index := scheduledTickIndex{pos: t.pos, hash: t.bhash}
		if existing, ok := queue.furthestTicks[index]; ok {
			queue.furthestTicks[index] = max(existing, t.t)
		} else {
			queue.furthestTicks[index] = t.t
		}
	}
}
