package item

// Flint is an item dropped rarely by gravel.
type Flint struct{}

func (Flint) EncodeItem() (name string, meta int16) {
	return "minecraft:flint", 0
}
