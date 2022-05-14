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

// String ...
func (Savanna) String() string {
	return "savanna"
}

// EncodeBiome ...
func (Savanna) EncodeBiome() int {
	return 35
}
