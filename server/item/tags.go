package item

// Tagged represents an item that has one or more item tags. These tags may be used by the client for various
// purposes, such as determining the tier of an item or checking if an item is food.
type Tagged interface {
	// Tags returns the tags of the item.
	Tags() []string
}
