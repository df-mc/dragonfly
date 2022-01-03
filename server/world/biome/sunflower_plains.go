package biome

// SunflowerPlains ...
type SunflowerPlains struct{}

// Temperature ...
func (SunflowerPlains) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (SunflowerPlains) Rainfall() float64 {
	return 0.4
}

// String ...
func (SunflowerPlains) String() string {
	return "sunflower_plains"
}

// EncodeBiome ...
func (SunflowerPlains) EncodeBiome() int {
	return 129
}
