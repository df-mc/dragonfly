package item

// PoppedChorusFruit is an item obtained by smelting chorus fruit, and used to craft end rods and purpur blocks.
// Unlike raw chorus fruit, the popped fruit is inedible.
type PoppedChorusFruit struct{}

// EncodeItem ...
func (PoppedChorusFruit) EncodeItem() (name string, meta int16) {
	return "minecraft:popped_chorus_fruit", 0
}
