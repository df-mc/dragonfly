package biome

// Grove ...
type Grove struct{}

// Temperature ...
func (Grove) Temperature() float64 {
	return -0.2
}

// Rainfall ...
func (Grove) Rainfall() float64 {
	return 0.8
}

// Ash ...
func (Grove) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (Grove) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (Grove) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (Grove) RedSpores() float64 {
	return 0
}

// String ...
func (Grove) String() string {
	return "grove"
}

// EncodeBiome ...
func (Grove) EncodeBiome() int {
	return 185
}
