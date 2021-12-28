package biome

// StonyShore ...
type StonyShore struct{}

// Temperature ...
func (StonyShore) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (StonyShore) Rainfall() float64 {
	return 0.3
}

// Ash ...
func (StonyShore) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (StonyShore) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (StonyShore) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (StonyShore) RedSpores() float64 {
	return 0
}

// String ...
func (StonyShore) String() string {
	return "stone_beach"
}

// EncodeBiome ...
func (StonyShore) EncodeBiome() int {
	return 25
}
