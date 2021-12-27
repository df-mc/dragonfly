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

// String ...
func (OldGrowthBirchForest) String() string {
	return "Old Growth Birch Forest"
}

// EncodeBiome ...
func (OldGrowthBirchForest) EncodeBiome() int {
	return 155
}
