package biome

// SnowyPlains ...
type SnowyPlains struct{}

// Temperature ...
func (SnowyPlains) Temperature() float64 {
	return 0
}

// Rainfall ...
func (SnowyPlains) Rainfall() float64 {
	return 0.5
}

// String ...
func (SnowyPlains) String() string {
	return "Snowy Plains"
}

// EncodeBiome ...
func (SnowyPlains) EncodeBiome() int {
	return 12
}
