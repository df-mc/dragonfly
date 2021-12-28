package biome

// Ocean ...
type Ocean struct{}

// Temperature ...
func (Ocean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (Ocean) Rainfall() float64 {
	return 0.5
}

// String ...
func (Ocean) String() string {
	return "ocean"
}

// EncodeBiome ...
func (Ocean) EncodeBiome() int {
	return 0
}
