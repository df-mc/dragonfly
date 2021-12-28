package biome

// Forest ...
type Forest struct{}

// Temperature ...
func (Forest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (Forest) Rainfall() float64 {
	return 0.8
}

// Ash ...
func (Forest) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Forest) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Forest) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Forest) RedSpores() float64 {
	return 0
}

// String ...
func (Forest) String() string {
	return "forest"
}

// EncodeBiome ...
func (Forest) EncodeBiome() int {
	return 4
}
