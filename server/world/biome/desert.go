package biome

// Desert ...
type Desert struct{}

// Temperature ...
func (Desert) Temperature() float64 {
	return 2
}

// Rainfall ...
func (Desert) Rainfall() float64 {
	return 0
}

// String ...
func (Desert) String() string {
	return "desert"
}

// EncodeBiome ...
func (Desert) EncodeBiome() int {
	return 2
}
