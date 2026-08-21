package item

// Digger represents an item configured as a digging tool, allowing it to break specific blocks
// faster than normal.
type Digger interface {
	// DiggerInfo returns the digger information of the item.
	DiggerInfo() DiggerInfo
}

// DiggerInfo is a struct returned by items that implement Digger. It contains the block-specific
// mining speed configuration of the item.
type DiggerInfo struct {
	// DestroySpeeds is a list of block-specific mining speed multipliers.
	DestroySpeeds []DestroySpeed
	// UseEfficiency specifies if the Efficiency enchantment will increase the dig speed of this
	// item.
	UseEfficiency bool
}

// DestroySpeed associates a block with a custom digging speed multiplier.
type DestroySpeed struct {
	// Block is the identifier of the block that can be dug.
	Block string
	// Speed is the digging speed multiplier for the correlating block.
	Speed float64
}
