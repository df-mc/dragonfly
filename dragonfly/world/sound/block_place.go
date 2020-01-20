package sound

import "github.com/dragonfly-tech/dragonfly/dragonfly/block"

// BlockPlace is a sound sent when a block is placed.
type BlockPlace struct {
	// Block is the block which is placed, for which a sound should be played. The sound played depends on
	// the block type.
	Block block.Block
}

func (BlockPlace) __() {}
