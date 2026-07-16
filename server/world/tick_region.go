package world

import (
	"fmt"
	"math/rand/v2"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/block/cube"
)

// This file implements region-based ticking, inspired by Folia's threaded
// regions: loaded chunks are partitioned into clusters that are more than
// 8 chunks apart. Each tick, the read-only scan for block tick candidates
// runs over these regions on worker goroutines while the owner goroutine
// waits. All block callbacks then run on the owner as usual, so the
// transaction model is unaffected.

const (
	// regionCellShift is the power of two of the width, in chunks, of the grid
	// cells used to cluster chunks: chunks in the same or adjacent cells always
	// share a region.
	regionCellShift = 3
	// regionBatchSize is the maximum number of chunks per gather job, so that
	// one big region still spreads across all tick workers.
	regionBatchSize = 64
)

// regionCell is the position of a region grid cell.
type regionCell [2]int32

// cellOf returns the region grid cell containing a chunk position.
func cellOf(pos ChunkPos) regionCell {
	return regionCell{pos[0] >> regionCellShift, pos[1] >> regionCellShift}
}

// cmpChunkPos orders chunk positions by x, then z.
func cmpChunkPos(a, b ChunkPos) int {
	if a[0] != b[0] {
		return int(a[0]) - int(b[0])
	}
	return int(a[1]) - int(b[1])
}

// tickRegion is a cluster of loaded chunks spatially independent of all other
// regions of the same World.
type tickRegion struct {
	chunks []regionChunk // sorted by position
	// min and max hold the region's bounding box in chunk coordinates. It may
	// overestimate after chunk removals, which only makes range checks more
	// conservative.
	min, max ChunkPos
}

// regionChunk is a single loaded chunk within a tickRegion.
type regionChunk struct {
	pos ChunkPos
	col *Column
}

// insert adds a chunk to the region, keeping chunks sorted and the bounding
// box up to date.
func (reg *tickRegion) insert(pos ChunkPos, col *Column) {
	i, _ := slices.BinarySearchFunc(reg.chunks, pos, func(rc regionChunk, p ChunkPos) int { return cmpChunkPos(rc.pos, p) })
	reg.chunks = slices.Insert(reg.chunks, i, regionChunk{pos: pos, col: col})
	reg.min[0], reg.min[1] = min(reg.min[0], pos[0]), min(reg.min[1], pos[1])
	reg.max[0], reg.max[1] = max(reg.max[0], pos[0]), max(reg.max[1], pos[1])
}

// remove deletes a chunk from the region. The bounding box is left unchanged.
func (reg *tickRegion) remove(pos ChunkPos) {
	if i, ok := slices.BinarySearchFunc(reg.chunks, pos, func(rc regionChunk, p ChunkPos) int { return cmpChunkPos(rc.pos, p) }); ok {
		reg.chunks = slices.Delete(reg.chunks, i, i+1)
	}
}

// Range classifications of a region relative to the loaders' positions.
const (
	regionOut = iota
	regionPartial
	regionIn
)

// rangeState reports whether the region is fully outside simulation distance
// r of all loaded positions, fully inside it for at least one, or partial.
func (reg *tickRegion) rangeState(loaded []ChunkPos, r int32) int {
	state := regionOut
	for _, pos := range loaded {
		dx := max(reg.min[0]-pos[0], 0, pos[0]-reg.max[0])
		dz := max(reg.min[1]-pos[1], 0, pos[1]-reg.max[1])
		if dx*dx+dz*dz > r*r {
			continue
		}
		// Within range of the near side; if the far corner is in range too,
		// every chunk of the region is within distance of this loader.
		fx := max(pos[0]-reg.min[0], reg.max[0]-pos[0])
		fz := max(pos[1]-reg.min[1], reg.max[1]-pos[1])
		if fx*fx+fz*fz <= r*r {
			return regionIn
		}
		state = regionPartial
	}
	return state
}

// regionPartition maintains the division of a World's loaded chunks into tick
// regions, updated incrementally as chunks load and unload. It must only be
// used from the World's owner goroutine.
type regionPartition struct {
	regions []*tickRegion // sorted by first chunk position
	cells   map[regionCell]*tickRegion
	counts  map[regionCell]int
	total   int
	dirty   bool
}

// markDirty schedules a full rebuild and releases all cached chunk references
// so unloaded chunks are not kept in memory until the next tick.
func (p *regionPartition) markDirty() {
	p.regions, p.cells, p.counts, p.total, p.dirty = nil, nil, nil, 0, true
}

// addChunk inserts a loaded chunk into the partition. A chunk in a cell that
// borders multiple regions merges them; a chunk in a free-standing cell forms
// a new region.
func (p *regionPartition) addChunk(pos ChunkPos, col *Column) {
	if p.dirty {
		return
	}
	if p.cells == nil {
		p.markDirty()
		return
	}
	cell := cellOf(pos)
	reg, ok := p.cells[cell]
	if !ok {
		var neighbours []*tickRegion
		for dx := int32(-1); dx <= 1; dx++ {
			for dz := int32(-1); dz <= 1; dz++ {
				if n, ok := p.cells[regionCell{cell[0] + dx, cell[1] + dz}]; ok && !slices.Contains(neighbours, n) {
					neighbours = append(neighbours, n)
				}
			}
		}
		switch len(neighbours) {
		case 0:
			reg = &tickRegion{min: pos, max: pos}
			p.regions = append(p.regions, reg)
		case 1:
			reg = neighbours[0]
		default:
			reg = p.merge(neighbours)
		}
		p.cells[cell] = reg
	}
	reg.insert(pos, col)
	p.counts[cell]++
	p.total++
	p.sortRegions()
}

// removeChunk removes an unloaded chunk from the partition. Emptying a cell
// may split a region, so that case falls back to a full rebuild.
func (p *regionPartition) removeChunk(pos ChunkPos) {
	if p.dirty {
		return
	}
	cell := cellOf(pos)
	reg, ok := p.cells[cell]
	if !ok || p.counts[cell] == 1 {
		p.markDirty()
		return
	}
	reg.remove(pos)
	p.counts[cell]--
	p.total--
	p.sortRegions()
}

// merge joins multiple regions into the first one, remapping their cells.
func (p *regionPartition) merge(regions []*tickRegion) *tickRegion {
	target := regions[0]
	for _, reg := range regions[1:] {
		target.chunks = append(target.chunks, reg.chunks...)
		target.min[0], target.min[1] = min(target.min[0], reg.min[0]), min(target.min[1], reg.min[1])
		target.max[0], target.max[1] = max(target.max[0], reg.max[0]), max(target.max[1], reg.max[1])
	}
	slices.SortFunc(target.chunks, func(a, b regionChunk) int { return cmpChunkPos(a.pos, b.pos) })
	for cell, reg := range p.cells {
		if slices.Contains(regions[1:], reg) {
			p.cells[cell] = target
		}
	}
	p.regions = slices.DeleteFunc(p.regions, func(reg *tickRegion) bool { return slices.Contains(regions[1:], reg) })
	return target
}

// sortRegions restores the deterministic region order. Chunk positions are
// unique, so regions' first chunks give a total order.
func (p *regionPartition) sortRegions() {
	slices.SortFunc(p.regions, func(a, b *tickRegion) int { return cmpChunkPos(a.chunks[0].pos, b.chunks[0].pos) })
}

// all returns the tick regions for the chunks passed, rebuilding the
// partition if it is dirty or inconsistent with the chunk map.
func (p *regionPartition) all(chunks map[ChunkPos]*Column) []*tickRegion {
	if p.dirty || p.cells == nil || p.total != len(chunks) {
		p.rebuild(chunks)
	}
	return p.regions
}

// rebuild recomputes the partition from scratch with a union-find over the
// region cell grid, merging cells that touch, including diagonally.
func (p *regionPartition) rebuild(chunks map[ChunkPos]*Column) {
	parent := make(map[regionCell]regionCell)
	var find func(c regionCell) regionCell
	find = func(c regionCell) regionCell {
		pr := parent[c]
		if pr == c {
			return c
		}
		root := find(pr)
		parent[c] = root
		return root
	}

	counts := make(map[regionCell]int)
	for pos := range chunks {
		cell := cellOf(pos)
		counts[cell]++
		if _, ok := parent[cell]; !ok {
			parent[cell] = cell
		}
	}
	for cell := range parent {
		for dx := int32(-1); dx <= 1; dx++ {
			for dz := int32(-1); dz <= 1; dz++ {
				n := regionCell{cell[0] + dx, cell[1] + dz}
				if _, ok := parent[n]; ok {
					if ra, rb := find(cell), find(n); ra != rb {
						parent[ra] = rb
					}
				}
			}
		}
	}

	roots := make(map[regionCell]*tickRegion)
	cells := make(map[regionCell]*tickRegion, len(parent))
	var regions []*tickRegion
	for pos, col := range chunks {
		cell := cellOf(pos)
		root := find(cell)
		reg, ok := roots[root]
		if !ok {
			reg = &tickRegion{min: pos, max: pos}
			roots[root] = reg
			regions = append(regions, reg)
		}
		cells[cell] = reg
		reg.min[0], reg.min[1] = min(reg.min[0], pos[0]), min(reg.min[1], pos[1])
		reg.max[0], reg.max[1] = max(reg.max[0], pos[0]), max(reg.max[1], pos[1])
		reg.chunks = append(reg.chunks, regionChunk{pos: pos, col: col})
	}
	for _, reg := range regions {
		slices.SortFunc(reg.chunks, func(a, b regionChunk) int { return cmpChunkPos(a.pos, b.pos) })
	}

	p.regions, p.cells, p.counts, p.total, p.dirty = regions, cells, counts, len(chunks), false
	p.sortRegions()
}

// tickBatch is one unit of gather work: a slice of one region's chunks with a
// batch-local random source and the output buffers filled by its worker.
// Batches are owned by the World and reused across ticks.
type tickBatch struct {
	chunks     []regionChunk
	pcg        rand.PCG
	checkRange bool

	blockEntities []cube.Pos
	randomBlocks  []cube.Pos
}

// workerPanic wraps a panic that occurred on a tick worker together with the
// worker's stack trace, so re-raising it on the owner stays diagnosable.
type workerPanic struct {
	value any
	stack []byte
}

func (p *workerPanic) Error() string {
	return fmt.Sprintf("tick worker: %v\n\nworker goroutine stack:\n%s", p.value, p.stack)
}

// runTickJobs runs fn for every batch on up to workers goroutines and blocks
// until all have completed. The workers only synchronise through a WaitGroup,
// so the call cannot deadlock; a worker panic is re-raised on the caller.
func runTickJobs(workers int, batches []tickBatch, fn func(b *tickBatch)) {
	if workers > len(batches) {
		workers = len(batches)
	}
	if workers <= 1 {
		for i := range batches {
			fn(&batches[i])
		}
		return
	}
	var (
		next     atomic.Int64
		panicked atomic.Pointer[workerPanic]
		wg       sync.WaitGroup
	)
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			defer func() {
				if v := recover(); v != nil {
					panicked.CompareAndSwap(nil, &workerPanic{value: v, stack: debug.Stack()})
				}
			}()
			for {
				i := int(next.Add(1)) - 1
				if i >= len(batches) {
					return
				}
				fn(&batches[i])
			}
		}()
	}
	wg.Wait()
	if v := panicked.Load(); v != nil {
		panic(v)
	}
}
