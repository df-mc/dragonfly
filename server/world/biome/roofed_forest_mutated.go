package biome

// DarkForestHills ...
type DarkForestHills struct{}

// Temperature ...
func (DarkForestHills) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (DarkForestHills) Rainfall() float64 {
	return 0.8
}

// String ...
func (DarkForestHills) String() string {
	return "Dark Forest Hills"
}

// EncodeBiome ...
func (DarkForestHills) EncodeBiome() int {
	return 157
}
