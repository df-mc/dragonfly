package biome

// OldGrowthPineTaiga ...
type OldGrowthPineTaiga struct{}

// Temperature ...
func (OldGrowthPineTaiga) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (OldGrowthPineTaiga) Rainfall() float64 {
	return 0.8
}

// String ...
func (OldGrowthPineTaiga) String() string {
	return "mega_taiga"
}

// EncodeBiome ...
func (OldGrowthPineTaiga) EncodeBiome() int {
	return 32
}
