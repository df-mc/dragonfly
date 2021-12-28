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

// Ash ...
func (SnowyBeach) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SnowyBeach) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SnowyBeach) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SnowyBeach) RedSpores() float64 {
	return 0
}

// String ...
func (SnowyBeach) String() string {
	return "cold_beach"
}

// EncodeBiome ...
func (SnowyBeach) EncodeBiome() int {
	return 26
}
