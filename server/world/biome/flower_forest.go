package biome

// FlowerForest ...
type FlowerForest struct{}

// Temperature ...
func (FlowerForest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (FlowerForest) Rainfall() float64 {
	return 0.8
}

// Ash ...
func (FlowerForest) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (FlowerForest) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (FlowerForest) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (FlowerForest) RedSpores() float64 {
	return 0
}

// String ...
func (FlowerForest) String() string {
	return "flower_forest"
}

// EncodeBiome ...
func (FlowerForest) EncodeBiome() int {
	return 132
}
