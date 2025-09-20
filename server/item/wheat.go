package item

// Wheat is a crop used to craft bread, cake, & cookies.
type Wheat struct{}

func (Wheat) CompostChance() float64 {
	return 0.65
}

func (w Wheat) EncodeItem() (name string, meta int16) {
	return "minecraft:wheat", 0
}
