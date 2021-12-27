package biome

// DeepOcean ...
type DeepOcean struct{}

// Temperature ...
func (DeepOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (DeepOcean) String() string {
	return "Deep Ocean"
}

// EncodeBiome ...
func (DeepOcean) EncodeBiome() int {
	return 24
}
