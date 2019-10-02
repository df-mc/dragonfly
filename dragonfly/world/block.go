package world

// Block is a block that may be placed or found in a world. In addition, the block may also be added to an
// inventory: It is also an item.
type Block interface {
	// Name returns the readable name of the block. An example for oak log would be 'Oak Log'.
	Name() string
}
