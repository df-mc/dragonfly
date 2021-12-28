package biome

// Meadow ...
type Meadow struct{}

// Temperature ...
func (Meadow) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (Meadow) Rainfall() float64 {
	return 0.8
}

// Ash ...
func (Meadow) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Meadow) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Meadow) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Meadow) RedSpores() float64 {
	return 0
}

// String ...
func (Meadow) String() string {
	return "meadow"
}

// EncodeBiome ...
func (Meadow) EncodeBiome() int {
	return 186
}
