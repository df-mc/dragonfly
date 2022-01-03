package biome

// DeepLukewarmOcean ...
type DeepLukewarmOcean struct{}

// Temperature ...
func (DeepLukewarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepLukewarmOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (DeepLukewarmOcean) String() string {
	return "deep_lukewarm_ocean"
}

// EncodeBiome ...
func (DeepLukewarmOcean) EncodeBiome() int {
	return 44
}
