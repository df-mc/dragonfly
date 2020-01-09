package item

// Item represents an item that may be added to an inventory.
type Item interface {
	// EncodeItem encodes an item to its Minecraft representation - A numerical ID with a numerical meta
	// value.
	EncodeItem() (id int32, meta int16)
}
