package item

// BlazePowder is an item made from a blaze rod obtained from blazes.
type BlazePowder struct{}

func (BlazePowder) EncodeItem() (name string, meta int16) {
	return "minecraft:blaze_powder", 0
}
