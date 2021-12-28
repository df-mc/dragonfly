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

// Ash ...
func (OldGrowthPineTaiga) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (OldGrowthPineTaiga) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (OldGrowthPineTaiga) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (OldGrowthPineTaiga) RedSpores() float64 {
	return 0
}

// String ...
func (OldGrowthPineTaiga) String() string {
	return "mega_taiga"
}

// EncodeBiome ...
func (OldGrowthPineTaiga) EncodeBiome() int {
	return 32
}
