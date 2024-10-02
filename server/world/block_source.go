package world

import "github.com/df-mc/dragonfly/server/block/cube"

// BlockSource represents a source for blocks. Blocks can be retrieved
// and set in the block source.
type BlockSource interface {
	// Block returns the block at the given position in the block source.
	Block(cube.Pos) Block
	// SetBlock sets the block at the given position in the block source.
	SetBlock(cube.Pos, Block, *SetOpts)
}
