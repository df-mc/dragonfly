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

// Ash ...
func (SunflowerPlains) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SunflowerPlains) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SunflowerPlains) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SunflowerPlains) RedSpores() float64 {
	return 0
}

// String ...
func (SunflowerPlains) String() string {
	return "sunflower_plains"
}

// EncodeBiome ...
func (SunflowerPlains) EncodeBiome() int {
	return 129
}
