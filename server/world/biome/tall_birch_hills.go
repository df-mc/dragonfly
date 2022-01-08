package biome

// TallBirchHills ...
type TallBirchHills struct{}

// Temperature ...
func (TallBirchHills) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (TallBirchHills) Rainfall() float64 {
	return 0.8
}

// String ...
func (TallBirchHills) String() string {
	return "birch_forest_hills_mutated"
}

// EncodeBiome ...
func (TallBirchHills) EncodeBiome() int {
	return 156
}
