package biome

// SnowyBeach ...
type SnowyBeach struct{}

// Temperature ...
func (SnowyBeach) Temperature() float64 {
	return 0.05
}

// Rainfall ...
func (SnowyBeach) Rainfall() float64 {
	return 0.3
}

// String ...
func (SnowyBeach) String() string {
	return "cold_beach"
}

// EncodeBiome ...
func (SnowyBeach) EncodeBiome() int {
	return 26
}
