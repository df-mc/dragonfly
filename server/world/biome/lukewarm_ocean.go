package biome

// LukewarmOcean ...
type LukewarmOcean struct{}

// Temperature ...
func (LukewarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (LukewarmOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (LukewarmOcean) String() string {
	return "lukewarm_ocean"
}

// EncodeBiome ...
func (LukewarmOcean) EncodeBiome() int {
	return 41
}
