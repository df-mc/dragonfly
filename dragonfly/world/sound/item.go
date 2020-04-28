package sound

// ItemBreak is a sound played when an item in the inventory is broken, such as when a tool reaches 0
// durability and breaks.
type ItemBreak struct{}

// ItemUseOn is a sound played when a player uses its item on a block. An example of this is when a player
// uses a shovel to turn grass into grass path. Note that in these cases, the Block is actually the new block,
// not the old one.
type ItemUseOn struct {
	// Block is generally the block that was created by using the item on a block. The sound played differs
	// depending on this field.
	Block block
}

func (i ItemBreak) __() {}
func (i ItemUseOn) __() {}
