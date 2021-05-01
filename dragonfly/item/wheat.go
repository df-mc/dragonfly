package item

// Wheat is a crop used to craft bread, cake, & cookies.
type Wheat struct{}

// EncodeItem ...
func (w Wheat) EncodeItem() (id int32, name string, meta int16) {
	return 296, "minecraft:wheat", 0
}
