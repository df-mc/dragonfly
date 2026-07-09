package chunk

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

// DirtyFlags describes which parts of a Column have changed since it was last
// marked clean.
type DirtyFlags uint8

const (
	// DirtyBlocks indicates that one or more block storages changed.
	DirtyBlocks DirtyFlags = 1 << iota
	// DirtyBiomes indicates that one or more biome storages changed.
	DirtyBiomes
	// DirtyEntities indicates that the entity list changed.
	DirtyEntities
	// DirtyBlockEntities indicates that the block entity list changed.
	DirtyBlockEntities
	// DirtyScheduledBlocks indicates that scheduled block updates changed.
	DirtyScheduledBlocks

	// DirtyChunk includes all block and biome data in a Column.
	DirtyChunk = DirtyBlocks | DirtyBiomes
	// DirtyAll includes every persisted part of a Column.
	DirtyAll = DirtyChunk | DirtyEntities | DirtyBlockEntities | DirtyScheduledBlocks
)

// Has reports if all flags passed are set.
func (flags DirtyFlags) Has(check DirtyFlags) bool {
	return flags&check == check
}

// Column holds all persisted data for a chunk column.
type Column struct {
	Chunk           *Chunk
	Entities        []Entity
	BlockEntities   []BlockEntity
	Tick            int64
	ScheduledBlocks []ScheduledBlockUpdate

	dirty DirtyFlags
}

// Dirty reports if any persisted part of the column changed since it was last
// marked clean.
func (col *Column) Dirty() bool {
	return col.DirtyFlags() != 0
}

// DirtyFlags returns the persisted parts of the column that changed since it
// was last marked clean.
func (col *Column) DirtyFlags() DirtyFlags {
	if col == nil {
		return 0
	}
	flags := col.dirty
	if col.Chunk != nil {
		flags |= col.Chunk.DirtyFlags()
	}
	return flags
}

// MarkDirty marks the persisted parts passed as changed.
func (col *Column) MarkDirty(flags DirtyFlags) {
	if col == nil {
		return
	}
	if col.Chunk != nil {
		col.Chunk.markDirty(flags & DirtyChunk)
	}
	col.dirty |= flags &^ DirtyChunk
}

// MarkClean marks every persisted part of the column as clean.
func (col *Column) MarkClean() {
	if col == nil {
		return
	}
	if col.Chunk != nil {
		col.Chunk.MarkClean()
	}
	col.dirty = 0
}

// MarkCleanFlags marks the persisted parts passed as clean.
func (col *Column) MarkCleanFlags(flags DirtyFlags) {
	if col == nil {
		return
	}
	if col.Chunk != nil {
		col.Chunk.markClean(flags & DirtyChunk)
	}
	col.dirty &^= flags &^ DirtyChunk
}

// BlockEntity is the persisted NBT data for a block entity at a position.
type BlockEntity struct {
	Pos  cube.Pos
	Data map[string]any
}

// Entity is the persisted NBT data for an entity.
type Entity struct {
	ID   int64
	Data map[string]any
}

// ScheduledBlockUpdate is a persisted scheduled block tick.
type ScheduledBlockUpdate struct {
	Pos   cube.Pos
	Block uint32
	Tick  int64
}
