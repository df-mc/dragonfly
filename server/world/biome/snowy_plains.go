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

// Ash ...
func (SnowyPlains) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SnowyPlains) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SnowyPlains) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SnowyPlains) RedSpores() float64 {
	return 0
}

// String ...
func (SnowyPlains) String() string {
	return "ice_plains"
}

// EncodeBiome ...
func (SnowyPlains) EncodeBiome() int {
	return 12
}
