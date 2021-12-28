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

// Ash ...
func (TallBirchHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (TallBirchHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (TallBirchHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (TallBirchHills) RedSpores() float64 {
	return 0
}

// String ...
func (TallBirchHills) String() string {
	return "birch_forest_hills_mutated"
}

// EncodeBiome ...
func (TallBirchHills) EncodeBiome() int {
	return 156
}
