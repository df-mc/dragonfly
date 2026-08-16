package item

// InteractButton represents an item that shows an interact button in touch controls.
type InteractButton interface {
	// InteractButton returns the text displayed on the interact button. If true, the default
	// "Use Item" text will be used.
	InteractButton() string
}
