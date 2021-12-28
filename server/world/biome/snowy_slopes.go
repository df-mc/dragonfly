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

// Ash ...
func (SnowySlopes) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SnowySlopes) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SnowySlopes) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SnowySlopes) RedSpores() float64 {
	return 0
}

// String ...
func (SnowySlopes) String() string {
	return "snowy_slopes"
}

// EncodeBiome ...
func (SnowySlopes) EncodeBiome() int {
	return 184
}
