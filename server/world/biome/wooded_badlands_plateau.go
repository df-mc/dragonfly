package biome

// WoodedBadlandsPlateau ...
type WoodedBadlandsPlateau struct{}

// Temperature ...
func (WoodedBadlandsPlateau) Temperature() float64 {
	return 2
}

// Rainfall ...
func (WoodedBadlandsPlateau) Rainfall() float64 {
	return 0
}

// Ash ...
func (WoodedBadlandsPlateau) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (WoodedBadlandsPlateau) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (WoodedBadlandsPlateau) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (WoodedBadlandsPlateau) RedSpores() float64 {
	return 0
}

// String ...
func (WoodedBadlandsPlateau) String() string {
	return "mesa_plateau_stone"
}

// EncodeBiome ...
func (WoodedBadlandsPlateau) EncodeBiome() int {
	return 39
}
