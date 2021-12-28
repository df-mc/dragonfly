package biome

// TaigaMountains ...
type TaigaMountains struct{}

// Temperature ...
func (TaigaMountains) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (TaigaMountains) Rainfall() float64 {
	return 0.8
}

// Ash ...
func (TaigaMountains) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (TaigaMountains) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (TaigaMountains) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (TaigaMountains) RedSpores() float64 {
	return 0
}

// String ...
func (TaigaMountains) String() string {
	return "taiga_mutated"
}

// EncodeBiome ...
func (TaigaMountains) EncodeBiome() int {
	return 133
}
