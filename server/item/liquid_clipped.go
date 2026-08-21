package item

// LiquidClipped represents an item that interacts with liquid blocks on use.
type LiquidClipped interface {
	// LiquidClipped returns whether the item interacts with liquid blocks on use.
	LiquidClipped() bool
}
