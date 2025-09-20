package item

// RabbitHide is an item dropped by rabbits.
type RabbitHide struct{}

func (RabbitHide) EncodeItem() (name string, meta int16) {
	return "minecraft:rabbit_hide", 0
}
