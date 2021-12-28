package biome

// Swamp ...
type Swamp struct{}

// Temperature ...
func (Swamp) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (Swamp) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (Swamp) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Swamp) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Swamp) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Swamp) RedSpores() float64 {
	return 0
}

// String ...
func (Swamp) String() string {
	return "swampland"
}

// EncodeBiome ...
func (Swamp) EncodeBiome() int {
	return 6
}
