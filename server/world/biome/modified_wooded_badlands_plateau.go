package biome

// ModifiedWoodedBadlandsPlateau ...
type ModifiedWoodedBadlandsPlateau struct{}

// Temperature ...
func (ModifiedWoodedBadlandsPlateau) Temperature() float64 {
	return 2
}

// Rainfall ...
func (ModifiedWoodedBadlandsPlateau) Rainfall() float64 {
	return 0
}

// Ash ...
func (ModifiedWoodedBadlandsPlateau) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (ModifiedWoodedBadlandsPlateau) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (ModifiedWoodedBadlandsPlateau) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (ModifiedWoodedBadlandsPlateau) RedSpores() float64 {
	return 0
}

// String ...
func (ModifiedWoodedBadlandsPlateau) String() string {
	return "mesa_plateau_stone_mutated"
}

// EncodeBiome ...
func (ModifiedWoodedBadlandsPlateau) EncodeBiome() int {
	return 167
}
