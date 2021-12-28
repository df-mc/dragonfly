package biome

// OldGrowthBirchForest ...
type OldGrowthBirchForest struct{}

// Temperature ...
func (OldGrowthBirchForest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (OldGrowthBirchForest) Rainfall() float64 {
	return 0.8
}

// Ash ...
func (OldGrowthBirchForest) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (OldGrowthBirchForest) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (OldGrowthBirchForest) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (OldGrowthBirchForest) RedSpores() float64 {
	return 0
}

// String ...
func (OldGrowthBirchForest) String() string {
	return "birch_forest_mutated"
}

// EncodeBiome ...
func (OldGrowthBirchForest) EncodeBiome() int {
	return 155
}
