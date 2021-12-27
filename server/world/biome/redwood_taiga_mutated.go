package biome

// OldGrowthSpruceTaiga ...
type OldGrowthSpruceTaiga struct{}

// Temperature ...
func (OldGrowthSpruceTaiga) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (OldGrowthSpruceTaiga) Rainfall() float64 {
	return 0.8
}

// String ...
func (OldGrowthSpruceTaiga) String() string {
	return "Old Growth Spruce Taiga"
}

// EncodeBiome ...
func (OldGrowthSpruceTaiga) EncodeBiome() int {
	return 160
}
