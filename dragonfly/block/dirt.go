package block

type (
	// Dirt is a block found abundantly in most biomes under a layer of grass blocks at the top of the
	// Overworld.
	Dirt struct{}
	// CoarseDirt is a variation of Dirt that grass blocks won't spread on.
	CoarseDirt struct{}
)

func (Dirt) Name() string {
	return "Dirt"
}

func (CoarseDirt) Name() string {
	return "Coarse Dirt"
}
