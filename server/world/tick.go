package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"math/rand"
	"time"
)

// ticker implements World ticking methods. World embeds this struct, so any exported methods on ticker are exported
// methods on World.
type ticker struct{ w *World }

// tickLoop starts ticking the World 20 times every second, updating all entities, blocks and other features such as
// the time and weather of the world, as required.
func (t ticker) tickLoop() {
	tc := time.NewTicker(time.Second / 20)
	defer tc.Stop()

	t.w.running.Add(1)
	for {
		select {
		case <-tc.C:
			t.tick()
		case <-t.w.closing:
			// World is being closed: Stop ticking and get rid of a task.
			t.w.running.Done()
			return
		}
	}
}

// tick performs a tick on the World and updates the time, weather, blocks and entities that require updates.
func (t ticker) tick() {
	viewers, loaders := t.w.allViewers()

	t.w.set.Lock()
	if len(viewers) == 0 && t.w.set.CurrentTick != 0 {
		t.w.set.Unlock()
		return
	}
	if t.w.advance {
		t.w.set.CurrentTick++
		if t.w.set.TimeCycle {
			t.w.set.Time++
		}
		if t.w.set.WeatherCycle {
			t.w.advanceWeather()
		}
	}

	rain, thunder, tick, tim := t.w.set.Raining, t.w.set.Thundering && t.w.set.Raining, t.w.set.CurrentTick, int(t.w.set.Time)
	t.w.set.Unlock()

	if tick%20 == 0 {
		for _, viewer := range viewers {
			if t.w.conf.Dim.TimeCycle() {
				viewer.ViewTime(tim)
			}
			if t.w.conf.Dim.WeatherCycle() {
				viewer.ViewWeather(rain, thunder)
			}
		}
	}
	if thunder {
		t.w.tickLightning()
	}

	t.tickEntities(tick)
	t.tickBlocksRandomly(loaders, tick)
	t.tickScheduledBlocks(tick)
	t.performNeighbourUpdates()
}

// tickScheduledBlocks executes scheduled block updates in chunks that are currently loaded.
func (t ticker) tickScheduledBlocks(tick int64) {
	t.w.updateMu.Lock()
	positions := make([]cube.Pos, 0, len(t.w.scheduledUpdates)/4)
	for pos, scheduledTick := range t.w.scheduledUpdates {
		if scheduledTick <= tick {
			positions = append(positions, pos)
			delete(t.w.scheduledUpdates, pos)
		}
	}
	t.w.updateMu.Unlock()

	for _, pos := range positions {
		if ticker, ok := t.w.Block(pos).(ScheduledTicker); ok {
			ticker.ScheduledTick(pos, t.w, t.w.r)
		}
		if liquid, ok := t.w.additionalLiquid(pos); ok {
			if ticker, ok := liquid.(ScheduledTicker); ok {
				ticker.ScheduledTick(pos, t.w, t.w.r)
			}
		}
	}
}

// performNeighbourUpdates performs all block updates that came as a result of a neighbouring block being changed.
func (t ticker) performNeighbourUpdates() {
	t.w.updateMu.Lock()
	positions := slices.Clone(t.w.neighbourUpdates)
	t.w.neighbourUpdates = t.w.neighbourUpdates[:0]
	t.w.updateMu.Unlock()

	for _, update := range positions {
		pos, changedNeighbour := update.pos, update.neighbour
		if ticker, ok := t.w.Block(pos).(NeighbourUpdateTicker); ok {
			ticker.NeighbourUpdateTick(pos, changedNeighbour, t.w)
		}
		if liquid, ok := t.w.additionalLiquid(pos); ok {
			if ticker, ok := liquid.(NeighbourUpdateTicker); ok {
				ticker.NeighbourUpdateTick(pos, changedNeighbour, t.w)
			}
		}
	}
}

// tickBlocksRandomly executes random block ticks in each sub chunk in the world that has at least one viewer
// registered from the viewers passed.
func (t ticker) tickBlocksRandomly(loaders []*Loader, tick int64) {
	var (
		r             = int32(t.w.tickRange())
		g             randUint4
		blockEntities []cube.Pos
		randomBlocks  []cube.Pos
	)
	if r == 0 {
		// NOP if the simulation distance is 0.
		return
	}

	loaded := make([]ChunkPos, 0, len(loaders))
	for _, loader := range loaders {
		loaded = append(loaded, loader.pos)
	}

	t.w.chunkMu.Lock()
	for pos, c := range t.w.chunks {
		if !t.anyWithinDistance(pos, loaded, r) {
			// No loaders in this chunk that are within the simulation distance, so proceed to the next.
			continue
		}
		c.Lock()
		blockEntities = append(blockEntities, maps.Keys(c.e)...)

		cx, cz := int(pos[0]<<4), int(pos[1]<<4)

		// We generate up to j random positions for every sub chunk.
		x, y, z := g.uint4(t.w.r), g.uint4(t.w.r), g.uint4(t.w.r)
		for j := 0; j < t.w.conf.RandomTickSpeed; j++ {
			for i, sub := range c.Sub() {
				if sub.Empty() {
					// SubChunk is empty, so skip it right away.
					continue
				}
				// Generally we would want to make sure the block has its block entities, but provided blocks
				// with block entities are generally ticked already, we are safe to assume that blocks
				// implementing the RandomTicker don't rely on additional block entity data.
				if rid := sub.Layers()[0].At(x, y, z); randomTickBlocks[rid] {
					subY := (i + (t.w.Range().Min() >> 4)) << 4
					randomBlocks = append(randomBlocks, cube.Pos{cx + int(x), subY + int(y), cz + int(z)})

					// Only generate new coordinates if a tickable block was actually found. If not, we can just re-use
					// the coordinates for the next sub chunk.
					x, y, z = g.uint4(t.w.r), g.uint4(t.w.r), g.uint4(t.w.r)
				}
			}
		}
		c.Unlock()
	}
	t.w.chunkMu.Unlock()

	for _, pos := range randomBlocks {
		if rb, ok := t.w.Block(pos).(RandomTicker); ok {
			rb.RandomTick(pos, t.w, t.w.r)
		}
	}
	for _, pos := range blockEntities {
		if tb, ok := t.w.Block(pos).(TickerBlock); ok {
			tb.Tick(tick, pos, t.w)
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
func (t ticker) tickEntities(tick int64) {
	type entityToMove struct {
		e             Entity
		after         *chunkData
		viewersBefore []Viewer
	}
	var (
		entitiesToMove []entityToMove
		entitiesToTick []TickerEntity
	)

	t.w.chunkMu.Lock()
	t.w.entityMu.Lock()
	for e, lastPos := range t.w.entities {
		chunkPos := chunkPosFromVec3(e.Position())

		c, ok := t.w.chunks[chunkPos]
		if !ok {
			continue
		}

		c.Lock()
		v := len(c.v)
		c.Unlock()

		if v > 0 {
			if ticker, ok := e.(TickerEntity); ok {
				entitiesToTick = append(entitiesToTick, ticker)
			}
		}

		if lastPos != chunkPos {
			// The entity was stored using an outdated chunk position. We update it and make sure it is ready
			// for loaders to view it.
			t.w.entities[e] = chunkPos
			var viewers []Viewer

			// When changing an entity's world, then teleporting it immediately, we could end up in a situation
			// where the old chunk of the entity was not loaded. In this case, it should be safe simply to ignore
			// the loaders from the old chunk. We can assume they never saw the entity in the first place.
			if old, ok := t.w.chunks[lastPos]; ok {
				old.Lock()
				old.entities = sliceutil.DeleteVal(old.entities, e)
				viewers = slices.Clone(old.v)
				old.Unlock()
			}
			entitiesToMove = append(entitiesToMove, entityToMove{e: e, viewersBefore: viewers, after: c})
		}
	}
	t.w.entityMu.Unlock()
	t.w.chunkMu.Unlock()

	for _, move := range entitiesToMove {
		move.after.Lock()
		move.after.entities = append(move.after.entities, move.e)
		viewersAfter := move.after.v
		move.after.Unlock()

		for _, viewer := range move.viewersBefore {
			if sliceutil.Index(viewersAfter, viewer) == -1 {
				// First we hide the entity from all loaders that were previously viewing it, but no
				// longer are.
				viewer.HideEntity(move.e)
			}
		}
		for _, viewer := range viewersAfter {
			if sliceutil.Index(move.viewersBefore, viewer) == -1 {
				// Then we show the entity to all loaders that are now viewing the entity in the new
				// chunk.
				showEntity(move.e, viewer)
			}
		}
	}
	for _, ticker := range entitiesToTick {
		// We gather entities to ticker and ticker them later, so that the lock on the entity mutex is no longer
		// active.
		ticker.Tick(t.w, tick)
	}
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
