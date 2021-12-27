package biome

// SnowyTaiga ...
type SnowyTaiga struct{}

// Temperature ...
func (SnowyTaiga) Temperature() float64 {
	return -0.5
}

// Rainfall ...
func (SnowyTaiga) Rainfall() float64 {
	return 0.4
}

// String ...
func (SnowyTaiga) String() string {
	return "Snowy Taiga"
}

// EncodeBiome ...
func (SnowyTaiga) EncodeBiome() int {
	return 30
}
