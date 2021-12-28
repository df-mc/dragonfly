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
	return "redwood_taiga_mutated"
}

// EncodeBiome ...
func (OldGrowthSpruceTaiga) EncodeBiome() int {
	return 160
}
