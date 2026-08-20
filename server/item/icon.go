package item

// Icon represents an item with a custom icon texture. The icon textures are the keys from the
// resource_pack/textures/item_texture.json 'texture_data' object associated with the texture file.
type Icon interface {
	// IconTextures returns the textures used for the item's icon. The "default" key contains the actual icon
	// texture of the item. Additional keys may be used to specify armour trim textures and palettes.
	IconTextures() map[string]string
}
