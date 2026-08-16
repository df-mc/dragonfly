package item

// HoverTextColor represents an item with a custom hover text color.
type HoverTextColor interface {
	// HoverTextColor returns the color of the item name when hovering over it.
	HoverTextColor() string
}
