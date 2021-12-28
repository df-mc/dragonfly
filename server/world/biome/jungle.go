package biome

// Jungle ...
type Jungle struct{}

// Temperature ...
func (Jungle) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (Jungle) Rainfall() float64 {
	return 0.9
}

// Ash ...
func (Jungle) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Jungle) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Jungle) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Jungle) RedSpores() float64 {
	return 0
}

// String ...
func (Jungle) String() string {
	return "jungle"
}

// EncodeBiome ...
func (Jungle) EncodeBiome() int {
	return 21
}
