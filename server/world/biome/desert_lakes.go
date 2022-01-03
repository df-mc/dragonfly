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

// String ...
func (DesertLakes) String() string {
	return "desert_mutated"
}

// EncodeBiome ...
func (DesertLakes) EncodeBiome() int {
	return 130
}
