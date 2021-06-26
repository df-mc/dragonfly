package item

// RabbitFoot is a brewing item obtained from rabbits.
type RabbitFoot struct{}

// EncodeItem ...
func (RabbitFoot) EncodeItem() (name string, meta int16) {
	return "minecraft:rabbit_foot", 0
}
