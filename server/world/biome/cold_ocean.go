package biome

// ColdOcean ...
type ColdOcean struct{}

// Temperature ...
func (ColdOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (ColdOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (ColdOcean) String() string {
	return "cold_ocean"
}

// EncodeBiome ...
func (ColdOcean) EncodeBiome() int {
	return 42
}
