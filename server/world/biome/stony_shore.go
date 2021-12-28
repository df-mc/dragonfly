package biome

// StonyShore ...
type StonyShore struct{}

// Temperature ...
func (StonyShore) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (StonyShore) Rainfall() float64 {
	return 0.3
}

// String ...
func (StonyShore) String() string {
	return "stone_beach"
}

// EncodeBiome ...
func (StonyShore) EncodeBiome() int {
	return 25
}
