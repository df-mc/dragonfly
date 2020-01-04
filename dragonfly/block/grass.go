package block

// Grass blocks generate abundantly across the surface of the Overworld.
type Grass struct{}

func (Grass) Minecraft() (name string, properties map[string]interface{}) {
	return "minecraft:grass", nil
}

func (Grass) Name() string {
	return "Grass"
}
