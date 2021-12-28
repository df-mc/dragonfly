package biome

// SnowyMountains ...
type SnowyMountains struct{}

// Temperature ...
func (SnowyMountains) Temperature() float64 {
	return 0
}

// Rainfall ...
func (SnowyMountains) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (SnowyMountains) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (SnowyMountains) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (SnowyMountains) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (SnowyMountains) RedSpores() float64 {
	return 0
}

// String ...
func (SnowyMountains) String() string {
	return "ice_mountains"
}

// EncodeBiome ...
func (SnowyMountains) EncodeBiome() int {
	return 13
}
