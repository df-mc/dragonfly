package item

// Totem is an uncommon combat item that can save holders from death.
type Totem struct{}

// MaxCount always returns 1.
func (Totem) MaxCount() int {
	return 1
}

// EncodeItem ...
func (Totem) EncodeItem() (name string, meta int16) {
	return "minecraft:totem_of_undying", 0
}

// OffHand ...
func (Totem) OffHand() bool {
	return true
}
