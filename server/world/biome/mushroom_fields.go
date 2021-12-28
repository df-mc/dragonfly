package biome

// MushroomFields ...
type MushroomFields struct{}

// Temperature ...
func (MushroomFields) Temperature() float64 {
	return 0.9
}

// Rainfall ...
func (MushroomFields) Rainfall() float64 {
	return 1
}

// String ...
func (MushroomFields) String() string {
	return "mushroom_island"
}

// EncodeBiome ...
func (MushroomFields) EncodeBiome() int {
	return 14
}
