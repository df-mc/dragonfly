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

// Ash ...
func (DarkForestHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (DarkForestHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (DarkForestHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (DarkForestHills) RedSpores() float64 {
	return 0
}

// String ...
func (DarkForestHills) String() string {
	return "roofed_forest_mutated"
}

// EncodeBiome ...
func (DarkForestHills) EncodeBiome() int {
	return 157
}
