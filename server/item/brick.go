package item

// Brick is an item made from clay, and is used for making bricks and flower pots.
type Brick struct{}

// EncodeItem ...
func (b Brick) EncodeItem() (name string, meta int16) {
	return "minecraft:brick", 0
}
