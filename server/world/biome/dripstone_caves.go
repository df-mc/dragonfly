package biome

// DripstoneCaves ...
type DripstoneCaves struct{}

// Temperature ...
func (DripstoneCaves) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (DripstoneCaves) Rainfall() float64 {
	return 0
}

// String ...
func (DripstoneCaves) String() string {
	return "dripstone_caves"
}

// EncodeBiome ...
func (DripstoneCaves) EncodeBiome() int {
	return 188
}
