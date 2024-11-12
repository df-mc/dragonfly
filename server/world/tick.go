package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"golang.org/x/exp/maps"
	"math/rand"
	"time"
)

// ticker implements World ticking methods. World embeds this struct, so any exported methods on ticker are exported
// methods on World.
type ticker struct{}

// tickLoop starts ticking the World 20 times every second, updating all entities, blocks and other features such as
// the time and weather of the world, as required.
func (t ticker) tickLoop(w *World) {
	tc := time.NewTicker(time.Second / 20)
	defer tc.Stop()

	w.running.Add(1)
	for {
		select {
		case <-tc.C:
			<-w.Exec(t.tick)
		case <-w.closing:
			// World is being closed: Stop ticking and get rid of a task.
			w.running.Done()
			return
		}
	}
}

// tick performs a tick on the World and updates the time, weather, blocks and entities that require updates.
func (t ticker) tick(tx *Tx) {
	viewers, loaders := tx.World().allViewers() // ALL VIEWERS

	tx.World().set.Lock()
	if len(viewers) == 0 && tx.World().set.CurrentTick != 0 {
		tx.World().set.Unlock()
		return
	}
	if tx.World().advance {
		tx.World().set.CurrentTick++
		if tx.World().set.TimeCycle {
			tx.World().set.Time++
		}
		if tx.World().set.WeatherCycle {
			tx.World().advanceWeather()
		}
	}

	rain, thunder, tick, tim := tx.World().set.Raining, tx.World().set.Thundering && tx.World().set.Raining, tx.World().set.CurrentTick, int(tx.World().set.Time)
	tx.World().set.Unlock()

	if tick%20 == 0 {
		for _, viewer := range viewers {
			if tx.World().Dimension().TimeCycle() {
				viewer.ViewTime(tim)
			}
			if tx.World().Dimension().WeatherCycle() {
				viewer.ViewWeather(rain, thunder)
			}
		}
	}
	if thunder {
		tx.World().tickLightning(tx)
	}

	t.tickEntities(tx, tick)
	t.tickBlocksRandomly(tx, loaders, tick)
	t.tickScheduledBlocks(tx, tick)
	t.performNeighbourUpdates(tx)
}

// tickScheduledBlocks executes scheduled block updates in chunks that are currently loaded.
func (t ticker) tickScheduledBlocks(tx *Tx, tick int64) {
	positions := make([]cube.Pos, 0, len(tx.World().scheduledUpdates)/4)
	for pos, scheduledTick := range tx.World().scheduledUpdates {
		if scheduledTick <= tick {
			positions = append(positions, pos)
			delete(tx.World().scheduledUpdates, pos)
		}
	}

	for _, pos := range positions {
		if ticker, ok := tx.Block(pos).(ScheduledTicker); ok {
			ticker.ScheduledTick(pos, tx, tx.World().r)
		}
		if liquid, ok := tx.World().additionalLiquid(pos); ok {
			if ticker, ok := liquid.(ScheduledTicker); ok {
				ticker.ScheduledTick(pos, tx, tx.World().r)
			}
		}
	}
}

// performNeighbourUpdates performs all block updates that came as a result of a neighbouring block being changed.
func (t ticker) performNeighbourUpdates(tx *Tx) {
	for _, update := range tx.World().neighbourUpdates {
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
	clear(tx.World().neighbourUpdates)
	tx.World().neighbourUpdates = tx.World().neighbourUpdates[:0]
}

// tickBlocksRandomly executes random block ticks in each sub chunk in the world that has at least one viewer
// registered from the viewers passed.
func (t ticker) tickBlocksRandomly(tx *Tx, loaders []*Loader, tick int64) {
	var (
		r             = int32(tx.World().tickRange())
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
		loader.mu.RLock()
		pos := loader.pos
		loader.mu.RUnlock()

		loaded = append(loaded, pos)
	}

	for pos, c := range tx.World().chunks {
		if !t.anyWithinDistance(pos, loaded, r) {
			// No loaders in this chunk that are within the simulation distance, so proceed to the next.
			continue
		}
		blockEntities = append(blockEntities, maps.Keys(c.BlockEntities)...)

		cx, cz := int(pos[0]<<4), int(pos[1]<<4)

		// We generate up to j random positions for every sub chunk.
		for j := 0; j < tx.World().conf.RandomTickSpeed; j++ {
			x, y, z := g.uint4(tx.World().r), g.uint4(tx.World().r), g.uint4(tx.World().r)

			for i, sub := range c.Sub() {
				if sub.Empty() {
					// SubChunk is empty, so skip it right away.
					continue
				}
				// Generally we would want to make sure the block has its block entities, but provided blocks
				// with block entities are generally ticked already, we are safe to assume that blocks
				// implementing the RandomTicker don't rely on additional block entity data.
				if rid := sub.Layers()[0].At(x, y, z); randomTickBlocks[rid] {
					subY := (i + (tx.Range().Min() >> 4)) << 4
					randomBlocks = append(randomBlocks, cube.Pos{cx + int(x), subY + int(y), cz + int(z)})

					// Only generate new coordinates if a tickable block was actually found. If not, we can just re-use
					// the coordinates for the next sub chunk.
					x, y, z = g.uint4(tx.World().r), g.uint4(tx.World().r), g.uint4(tx.World().r)
				}
			}
		}
	}

	for _, pos := range randomBlocks {
		if rb, ok := tx.Block(pos).(RandomTicker); ok {
			rb.RandomTick(pos, tx, tx.World().r)
		}
	}
	for _, pos := range blockEntities {
		if tb, ok := tx.Block(pos).(TickerBlock); ok {
			tb.Tick(tick, pos, tx)
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
				if sliceutil.Index(c.viewers, viewer) == -1 {
					// First we hide the entity from all loaders that were previously viewing it, but no
					// longer are.
					viewer.HideEntity(e)
				}
			}
			for _, viewer := range c.viewers {
				if sliceutil.Index(viewers, viewer) == -1 {
					// Then we show the entity to all loaders that are now viewing the entity in the new
					// chunk.
					showEntity(e, viewer)
				}
			}
		}

		if len(c.viewers) > 0 {
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
