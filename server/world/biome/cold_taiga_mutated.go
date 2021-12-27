package biome

// SnowyTaigaMountains ...
type SnowyTaigaMountains struct{}

// Temperature ...
func (SnowyTaigaMountains) Temperature() float64 {
	return -0.5
}

// Rainfall ...
func (SnowyTaigaMountains) Rainfall() float64 {
	return 0.4
}

// String ...
func (SnowyTaigaMountains) String() string {
	return "Snowy Taiga Mountains"
}

// EncodeBiome ...
func (SnowyTaigaMountains) EncodeBiome() int {
	return 158
}
