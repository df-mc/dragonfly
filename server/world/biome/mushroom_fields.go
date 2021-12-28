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

// Ash ...
func (MushroomFields) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (MushroomFields) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (MushroomFields) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (MushroomFields) RedSpores() float64 {
	return 0
}

// String ...
func (MushroomFields) String() string {
	return "mushroom_island"
}

// EncodeBiome ...
func (MushroomFields) EncodeBiome() int {
	return 14
}
