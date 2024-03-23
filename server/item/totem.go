package item

type Totem struct{}

// MaxCount always returns 1.
func (Totem) MaxCount() int {
	return 1
}

// EncodeItem ...
func (Totem) EncodeItem() (name string, meta int16) {
	return "minecraft:totem_of_undying", 0
}
