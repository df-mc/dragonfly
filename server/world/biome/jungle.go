package biome

// Jungle ...
type Jungle struct{}

// Temperature ...
func (Jungle) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (Jungle) Rainfall() float64 {
	return 0.9
}

// String ...
func (Jungle) String() string {
	return "jungle"
}

// EncodeBiome ...
func (Jungle) EncodeBiome() int {
	return 21
}
