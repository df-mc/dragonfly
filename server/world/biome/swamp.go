package biome

// Swamp ...
type Swamp struct{}

// Temperature ...
func (Swamp) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (Swamp) Rainfall() float64 {
	return 0.5
}

// String ...
func (Swamp) String() string {
	return "swampland"
}

// EncodeBiome ...
func (Swamp) EncodeBiome() int {
	return 6
}
