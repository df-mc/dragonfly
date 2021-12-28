package biome

// LegacyFrozenOcean ...
type LegacyFrozenOcean struct{}

// Temperature ...
func (LegacyFrozenOcean) Temperature() float64 {
	return 0
}

// Rainfall ...
func (LegacyFrozenOcean) Rainfall() float64 {
	return 0.5
}

// String ...
func (LegacyFrozenOcean) String() string {
	return "legacy_frozen_ocean"
}

// EncodeBiome ...
func (LegacyFrozenOcean) EncodeBiome() int {
	return 47
}
