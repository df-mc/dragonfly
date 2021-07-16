package item

// Wheat is a crop used to craft bread, cake, & cookies.
type Wheat struct{}

// EncodeItem ...
func (w Wheat) EncodeItem() (name string, meta int16) {
	return "minecraft:wheat", 0
}
