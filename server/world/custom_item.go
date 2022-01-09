package world

import (
	"github.com/df-mc/dragonfly/server/item/category"
	"image"
)

// CustomItem represents an item that is non-vanilla and requires a resource pack and extra steps to show it
// to the client.
type CustomItem interface {
	Item
	// Name is the name that will be displayed on the item to all clients.
	Name() string
	// Texture is the Image of the texture for this item.
	Texture() image.Image
	// Category is the category the item will be listed under in the creative inventory.
	Category() category.Category
}

// customItems holds a list of all registered custom items.
var customItems []CustomItem

// CustomItems returns a slice of all registered custom items.
func CustomItems() []CustomItem {
	m := make([]CustomItem, 0, len(customItems))
	for _, i := range customItems {
		m = append(m, i)
	}
	return m
}
