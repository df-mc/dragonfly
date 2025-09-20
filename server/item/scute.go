package item

// Scute is an item that baby turtles drop when they grow into adults.
type Scute struct{}

func (Scute) EncodeItem() (name string, meta int16) {
	return "minecraft:turtle_scute", 0
}
