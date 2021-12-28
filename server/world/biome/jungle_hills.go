package biome

// JungleHills ...
type JungleHills struct{}

// Temperature ...
func (JungleHills) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (JungleHills) Rainfall() float64 {
	return 0.9
}

// Ash ...
func (JungleHills) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (JungleHills) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (JungleHills) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (JungleHills) RedSpores() float64 {
	return 0
}

// String ...
func (JungleHills) String() string {
	return "jungle_hills"
}

// EncodeBiome ...
func (JungleHills) EncodeBiome() int {
	return 22
}
