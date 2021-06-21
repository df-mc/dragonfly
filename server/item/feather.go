package item

// Feather are items dropped by chickens and parrots, as well as tamed cats as morning gifts.
type Feather struct{}

// EncodeItem ...
func (Feather) EncodeItem() (name string, meta int16) {
	return "minecraft:feather", 0
}
