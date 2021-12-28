package biome

// Ocean ...
type Ocean struct{}

// Temperature ...
func (Ocean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (Ocean) Rainfall() float64 {
	return 0.5
}

// Ash ...
func (Ocean) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Ocean) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Ocean) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Ocean) RedSpores() float64 {
	return 0
}

// String ...
func (Ocean) String() string {
	return "ocean"
}

// EncodeBiome ...
func (Ocean) EncodeBiome() int {
	return 0
}
