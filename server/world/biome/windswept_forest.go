package biome

// WindsweptForest ...
type WindsweptForest struct{}

// Temperature ...
func (WindsweptForest) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (WindsweptForest) Rainfall() float64 {
	return 0.3
}

// String ...
func (WindsweptForest) String() string {
	return "extreme_hills_plus_trees"
}

// EncodeBiome ...
func (WindsweptForest) EncodeBiome() int {
	return 34
}
