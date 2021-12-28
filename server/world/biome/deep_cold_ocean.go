package biome

// DeepColdOcean ...
type DeepColdOcean struct{}

// Temperature ...
func (DeepColdOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepColdOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (DeepColdOcean) String() string {
	return "deep_cold_ocean"
}

// EncodeBiome ...
func (DeepColdOcean) EncodeBiome() int {
	return 45
}
