package block

// Grass blocks generate abundantly across the surface of the Overworld.
type Grass struct{}

func (Grass) Name() string {
	return "Grass"
}
