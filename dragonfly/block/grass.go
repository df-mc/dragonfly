package block

// Grass blocks generate abundantly across the surface of the Overworld.
type Grass struct{}

func (Grass) EncodeItem() (id int32, meta int16) {
	return 2, 0
}

func (Grass) Minecraft() (name string, properties map[string]interface{}) {
	return "minecraft:grass", nil
}

func (Grass) Name() string {
	return "Grass"
}
