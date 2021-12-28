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

// Ash ...
func (LukewarmOcean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (LukewarmOcean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (LukewarmOcean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (LukewarmOcean) RedSpores() float64 {
	return 0
}

// String ...
func (LukewarmOcean) String() string {
	return "lukewarm_ocean"
}

// EncodeBiome ...
func (LukewarmOcean) EncodeBiome() int {
	return 41
}
