package particle

import "github.com/dragonfly-tech/dragonfly/dragonfly/block"

// BlockBreak is a particle sent when a block is broken. It represents a bunch of particles that are textured
// like the block that the particle holds.
type BlockBreak struct {
	// Block is the block of which particles should be shown. The particles will change depending on what
	// block is held.
	Block block.Block
}

func (BlockBreak) __() {}
