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

// Ash ...
func (DeepLukewarmOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (DeepLukewarmOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (DeepLukewarmOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (DeepLukewarmOcean) RedSpores() float64 {
	return 0
}

// String ...
func (DeepLukewarmOcean) String() string {
	return "deep_lukewarm_ocean"
}

// EncodeBiome ...
func (DeepLukewarmOcean) EncodeBiome() int {
	return 44
}
