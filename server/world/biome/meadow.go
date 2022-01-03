package biome

// Meadow ...
type Meadow struct{}

// Temperature ...
func (Meadow) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (Meadow) Rainfall() float64 {
	return 0.8
}

// String ...
func (Meadow) String() string {
	return "meadow"
}

// EncodeBiome ...
func (Meadow) EncodeBiome() int {
	return 186
}
