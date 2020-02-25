package block

// Grass blocks generate abundantly across the surface of the Overworld.
type Grass struct{}

// EncodeItem ...
func (Grass) EncodeItem() (id int32, meta int16) {
	return 2, 0
}

// EncodeBlock ...
func (Grass) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:grass", nil
}
