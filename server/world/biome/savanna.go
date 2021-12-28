package biome

// Savanna ...
type Savanna struct{}

// Temperature ...
func (Savanna) Temperature() float64 {
	return 1.2
}

// Rainfall ...
func (Savanna) Rainfall() float64 {
	return 0
}

// Ash ...
func (Savanna) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Savanna) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Savanna) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Savanna) RedSpores() float64 {
	return 0
}

// String ...
func (Savanna) String() string {
	return "savanna"
}

// EncodeBiome ...
func (Savanna) EncodeBiome() int {
	return 35
}
