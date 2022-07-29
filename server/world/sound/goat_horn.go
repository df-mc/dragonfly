package sound

// Horn represents a variant of a goat horn.
type Horn struct {
	goatHornType
}

// Ponder returns the 'Ponder' goat horn type.
func Ponder() Horn {
	return Horn{0}
}

// Sing returns the 'Sing' goat horn type.
func Sing() Horn {
	return Horn{1}
}

// Seek returns the 'Seek' goat horn type.
func Seek() Horn {
	return Horn{2}
}

// Feel returns the 'Feel' goat horn type.
func Feel() Horn {
	return Horn{3}
}

// Admire returns the 'Admire' goat horn type.
func Admire() Horn {
	return Horn{4}
}

// Call returns the 'Call' goat horn type.
func Call() Horn {
	return Horn{5}
}

// Yearn returns the 'Yearn' goat horn type.
func Yearn() Horn {
	return Horn{6}
}

// Dream returns the 'Dream' goat horn type.
func Dream() Horn {
	return Horn{7}
}

type goatHornType uint8

// Uint8 returns the goat horn type as a uint8.
func (g goatHornType) Uint8() uint8 {
	return uint8(g)
}

// Name returns the goat horn type's name.
func (g goatHornType) Name() string {
	switch g {
	case 0:
		return "Ponder"
	case 1:
		return "Sing"
	case 2:
		return "Seek"
	case 3:
		return "Feel"
	case 4:
		return "Admire"
	case 5:
		return "Call"
	case 6:
		return "Yearn"
	case 7:
		return "Dream"
	}
	panic("should never happen")
}

// GoatHorns ...
func GoatHorns() []Horn {
	return []Horn{Ponder(), Sing(), Seek(), Feel(), Admire(), Call(), Yearn(), Dream()}
}
