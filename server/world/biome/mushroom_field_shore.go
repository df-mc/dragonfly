package biome

// MushroomFieldShore ...
type MushroomFieldShore struct{}

// Temperature ...
func (MushroomFieldShore) Temperature() float64 {
	return 0.9
}

// Rainfall ...
func (MushroomFieldShore) Rainfall() float64 {
	return 1
}

// String ...
func (MushroomFieldShore) String() string {
	return "mushroom_island_shore"
}

// EncodeBiome ...
func (MushroomFieldShore) EncodeBiome() int {
	return 15
}
