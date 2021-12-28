package biome

// DesertLakes ...
type DesertLakes struct{}

// Temperature ...
func (DesertLakes) Temperature() float64 {
	return 2
}

// Rainfall ...
func (DesertLakes) Rainfall() float64 {
	return 0
}

// Ash ...
func (DesertLakes) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (DesertLakes) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (DesertLakes) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (DesertLakes) RedSpores() float64 {
	return 0
}

// String ...
func (DesertLakes) String() string {
	return "desert_mutated"
}

// EncodeBiome ...
func (DesertLakes) EncodeBiome() int {
	return 130
}
