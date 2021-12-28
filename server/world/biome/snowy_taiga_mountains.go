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

// Ash ...
func (SnowyTaigaMountains) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SnowyTaigaMountains) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SnowyTaigaMountains) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SnowyTaigaMountains) RedSpores() float64 {
	return 0
}

// String ...
func (SnowyTaigaMountains) String() string {
	return "cold_taiga_mutated"
}

// EncodeBiome ...
func (SnowyTaigaMountains) EncodeBiome() int {
	return 158
}
