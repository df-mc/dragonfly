package biome

// Badlands ...
type Badlands struct{}

// Temperature ...
func (Badlands) Temperature() float64 {
	return 2
}

// Rainfall ...
func (Badlands) Rainfall() float64 {
	return 0
}

// Ash ...
func (Badlands) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Badlands) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Badlands) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Badlands) RedSpores() float64 {
	return 0
}

// String ...
func (Badlands) String() string {
	return "mesa"
}

// EncodeBiome ...
func (Badlands) EncodeBiome() int {
	return 37
}
