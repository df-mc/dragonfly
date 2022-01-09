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

// String ...
func (SnowyMountains) String() string {
	return "ice_mountains"
}

// EncodeBiome ...
func (SnowyMountains) EncodeBiome() int {
	return 13
}
