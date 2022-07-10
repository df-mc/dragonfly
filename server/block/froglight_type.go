package block

// FroglightType represents a type of froglight.
type FroglightType struct {
	froglight
}

type froglight uint8

// Pearlescent is the purple variant of froglight.
func Pearlescent() FroglightType {
	return FroglightType{0}
}

// Verdant is the green variant of froglight.
func Verdant() FroglightType {
	return FroglightType{1}
}

// Ochre is the yellow variant of froglight.
func Ochre() FroglightType {
	return FroglightType{2}
}

// Uint8 ...
func (f froglight) Uint8() uint8 {
	return uint8(f)
}

// Name ...
func (f froglight) Name() string {
	switch f {
	case 0:
		return "Pearlescent Froglight"
	case 1:
		return "Verdant Froglight"
	case 2:
		return "Ochre Froglight"
	}
	panic("unknown froglight type")
}

// String ...
func (f froglight) String() string {
	switch f {
	case 0:
		return "pearlescent"
	case 1:
		return "verdant"
	case 2:
		return "ochre"
	}
	panic("unknown froglight type")
}

// FroglightTypes ...
func FroglightTypes() []FroglightType {
	return []FroglightType{Pearlescent(), Verdant(), Ochre()}
}
