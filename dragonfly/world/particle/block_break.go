package particle

// BlockBreak is a particle sent when a block is broken. It represents a bunch of particles that are textured
// like the block that the particle holds.
type BlockBreak struct {
	// Block is the block of which particles should be shown. The particles will change depending on what
	// block is held.
	Block block
}

type block interface {
	EncodeBlock() (name string, properties map[string]interface{})
}

func (BlockBreak) __() {}
