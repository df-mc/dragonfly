package biome

// DeepWarmOcean ...
type DeepWarmOcean struct{}

// Temperature ...
func (DeepWarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepWarmOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (DeepWarmOcean) String() string {
	return "deep_warm_ocean"
}

// EncodeBiome ...
func (DeepWarmOcean) EncodeBiome() int {
	return 43
}
