package item

// BlockPlacer represents an item that can place blocks when used.
type BlockPlacer interface {
	// BlockPlacerInfo returns the block placer information of the item.
	BlockPlacerInfo() BlockPlacerInfo
}

// BlockPlacerInfo is a struct returned by items that implement BlockPlacer. It contains the block
// placement configuration of the item.
type BlockPlacerInfo struct {
	// Block is the identifier of the block that will be placed.
	Block string
	// UseOn is a list of block identifiers that this item can be used on. If empty, all blocks
	// are allowed.
	UseOn []string
	// ReplaceBlockItem specifies if the item will be registered as the item for this block.
	ReplaceBlockItem bool
	// AlignedPlacement specifies if block placement through this item is aligned while the
	// interaction button is held down.
	AlignedPlacement bool
}
