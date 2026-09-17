package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

// loadedSource is a BlockSource reading from the chunks a World has loaded. Positions in an unloaded
// chunk or outside the World's Range read as air.
type loadedSource struct{ w *World }

// Block ...
func (s loadedSource) Block(pos cube.Pos) Block {
	b, _ := s.w.blockLoaded(pos)
	return b
}

// derivedBlock is a StateDeriver found in a chunk at a position.
type derivedBlock struct {
	pos cube.Pos
	b   StateDeriver
}

// deriveStates recalculates the derived states of the chunk at centre and of the chunks around it, so
// that a chunk that was missing neighbours earlier is upgraded once the chunk completing it is loaded.
func (w *World) deriveStates(centre ChunkPos) {
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			pos := ChunkPos{centre[0] + x, centre[1] + z}
			if c, ok := w.chunks[pos]; ok && c.LegacyStates() {
				w.deriveChunkStates(pos, c)
			}
		}
	}
}

// deriveChunkStates recalculates the derived states of the chunk at pos. Blocks on the edge of a chunk
// derive their state partly from the chunk next to it, so the chunk is left marked and untouched until
// every chunk around it is loaded.
func (w *World) deriveChunkStates(pos ChunkPos, c *Column) {
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			if _, ok := w.chunks[ChunkPos{pos[0] + x, pos[1] + z}]; !ok {
				return
			}
		}
	}
	found := w.derivedBlocks(pos, c)
	c.MarkStatesUpgraded()
	if len(found) == 0 {
		return
	}

	src := loadedSource{w: w}
	for _, f := range found {
		derived := f.b.DeriveState(f.pos, src)
		if derived == Block(f.b) {
			continue
		}
		// Written straight into the chunk rather than through Tx.SetBlock, which would run the
		// neighbour updates of a state that was only ever wrong in memory.
		c.SetBlock(uint8(f.pos[0]), int16(f.pos[1]), uint8(f.pos[2]), 0, w.conf.Blocks.BlockRuntimeID(derived))
		c.modified = true

		// A chunk whose neighbours arrived after it was sent is upgraded while viewers already hold
		// the state it was loaded with, so they are told about the block themselves.
		for _, v := range c.viewers {
			v.ViewBlockUpdate(f.pos, derived, 0)
		}
	}
}

// derivedBlocks returns every StateDeriver in the chunk at pos. Sub chunks are filtered on their
// palette first, so that one holding none of them is skipped without reading any block in it.
func (w *World) derivedBlocks(pos ChunkPos, c *Column) []derivedBlock {
	var found []derivedBlock
	baseX, baseZ := int(pos[0])<<4, int(pos[1])<<4
	for index, sub := range c.Sub() {
		if sub.Empty() {
			continue
		}
		storage := sub.Layer(0)
		palette := storage.Palette()

		// derivers maps palette indices holding a StateDeriver to that block, and stays nil if none do.
		var derivers map[uint16]StateDeriver
		for i := 0; i < palette.Len(); i++ {
			b, ok := w.conf.Blocks.BlockByRuntimeID(palette.Value(uint16(i)))
			if !ok {
				continue
			}
			if d, ok := b.(StateDeriver); ok {
				if derivers == nil {
					derivers = make(map[uint16]StateDeriver, 4)
				}
				derivers[uint16(i)] = d
			}
		}
		if derivers == nil {
			continue
		}

		baseY := int(c.SubY(int16(index)))
		for x := byte(0); x < 16; x++ {
			for y := byte(0); y < 16; y++ {
				for z := byte(0); z < 16; z++ {
					if d, ok := derivers[storage.PaletteIndex(x, y, z)]; ok {
						found = append(found, derivedBlock{
							pos: cube.Pos{baseX + int(x), baseY + int(y), baseZ + int(z)},
							b:   d,
						})
					}
				}
			}
		}
	}
	return found
}
