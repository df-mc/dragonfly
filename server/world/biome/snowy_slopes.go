package biome

// SnowySlopes ...
type SnowySlopes struct{}

// Temperature ...
func (SnowySlopes) Temperature() float64 {
	return -0.3
}

// Rainfall ...
func (SnowySlopes) Rainfall() float64 {
	return 0.9
}

// String ...
func (SnowySlopes) String() string {
	return "snowy_slopes"
}

// EncodeBiome ...
func (SnowySlopes) EncodeBiome() int {
	return 184
}
