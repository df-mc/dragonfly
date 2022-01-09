package biome

// LushCaves ...
type LushCaves struct{}

// Temperature ...
func (LushCaves) Temperature() float64 {
	return 0.9
}

// Rainfall ...
func (LushCaves) Rainfall() float64 {
	return 0
}

// String ...
func (LushCaves) String() string {
	return "lush_caves"
}

// EncodeBiome ...
func (LushCaves) EncodeBiome() int {
	return 187
}
