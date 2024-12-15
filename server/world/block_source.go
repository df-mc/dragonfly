package world

import "github.com/df-mc/dragonfly/server/block/cube"

// BlockSource represents a source for obtaining blocks.
type BlockSource interface {
	// Block returns the block at the given position in the block source.
	Block(cube.Pos) Block
}

// worldSource is a wrapper around a World that implements BlockSource.
type worldSource struct{ w *World }

func (w worldSource) Block(pos cube.Pos) Block { return w.w.block(pos) }
