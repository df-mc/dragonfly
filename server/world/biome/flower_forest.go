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

// String ...
func (FlowerForest) String() string {
	return "flower_forest"
}

// EncodeBiome ...
func (FlowerForest) EncodeBiome() int {
	return 132
}
