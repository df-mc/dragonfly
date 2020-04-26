package sound

// BlockPlace is a sound sent when a block is placed.
type BlockPlace struct {
	// Block is the block which is placed, for which a sound should be played. The sound played depends on
	// the block type.
	Block block

	sound
}

// BlockBreaking is a sound sent continuously while a player is breaking a block.
type BlockBreaking struct {
	// Block is the block which is being broken, for which a sound should be played. The sound played depends
	// on the block type.
	Block block

	sound
}

// ChestOpen is played when a chest is opened.
type ChestOpen struct{ sound }

// ChestClose is played when a chest is closed.
type ChestClose struct{ sound }

type block interface {
	EncodeBlock() (name string, properties map[string]interface{})
}

func (BlockPlace) __() {}
