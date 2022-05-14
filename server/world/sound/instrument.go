package sound

// Instrument represents a note block instrument.
type Instrument struct {
	instrument
}

type instrument int32

// Int32 ...
func (i instrument) Int32() int32 {
	return int32(i)
}

// Piano is an instrument type for the note block.
func Piano() Instrument {
	return Instrument{0}
}

// BassDrum is an instrument type for the note block.
func BassDrum() Instrument {
	return Instrument{1}
}

// Snare is an instrument type for the note block.
func Snare() Instrument {
	return Instrument{2}
}

// ClicksAndSticks is an instrument type for the note block.
func ClicksAndSticks() Instrument {
	return Instrument{3}
}

// Bass is an instrument type for the note block.
func Bass() Instrument {
	return Instrument{4}
}

// Bell is an instrument type for the note block.
func Bell() Instrument {
	return Instrument{5}
}

// Flute is an instrument type for the note block.
func Flute() Instrument {
	return Instrument{6}
}

// Chimes is an instrument type for the note block.
func Chimes() Instrument {
	return Instrument{7}
}

// Guitar is an instrument type for the note block.
func Guitar() Instrument {
	return Instrument{8}
}

// Xylophone is an instrument type for the note block.
func Xylophone() Instrument {
	return Instrument{9}
}

// IronXylophone is an instrument type for the note block.
func IronXylophone() Instrument {
	return Instrument{10}
}

// CowBell is an instrument type for the note block.
func CowBell() Instrument {
	return Instrument{11}
}

// Didgeridoo is an instrument type for the note block.
func Didgeridoo() Instrument {
	return Instrument{12}
}

// Bit is an instrument type for the note block.
func Bit() Instrument {
	return Instrument{13}
}

// Banjo is an instrument type for the note block.
func Banjo() Instrument {
	return Instrument{14}
}

// Pling is an instrument type for the note block.
func Pling() Instrument {
	return Instrument{15}
}
