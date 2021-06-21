package item

// RabbitHide is an item dropped by rabbits.
type RabbitHide struct{}

// EncodeItem ...
func (RabbitHide) EncodeItem() (name string, meta int16) {
	return "minecraft:rabbit_hide", 0
}
