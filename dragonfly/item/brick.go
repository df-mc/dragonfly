package item

// Brick is an item made from clay, and is used for making bricks and flower pots.
type Brick struct{}

// EncodeItem ...
func (b Brick) EncodeItem() (id int32, name string, meta int16) {
	return 336, "minecraft:brick", 0
}
