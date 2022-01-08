package biome

// WarmOcean ...
type WarmOcean struct{}

// Temperature ...
func (WarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (WarmOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (WarmOcean) String() string {
	return "warm_ocean"
}

// EncodeBiome ...
func (WarmOcean) EncodeBiome() int {
	return 40
}
