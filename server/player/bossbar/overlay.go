package bossbar

// Overlay is the overlay rendered on top of a BossBar. It controls whether the bar is
// shown as a solid bar or split into a number of notched segments.
type Overlay struct{ overlay }

// Progress is the overlay for a solid, continuous boss bar. It is the default overlay.
func Progress() Overlay {
	return Overlay{overlay(0)}
}

// Notched6 is the overlay for a boss bar split into 6 segments.
func Notched6() Overlay {
	return Overlay{overlay(1)}
}

// Notched10 is the overlay for a boss bar split into 10 segments.
func Notched10() Overlay {
	return Overlay{overlay(2)}
}

// Notched12 is the overlay for a boss bar split into 12 segments.
func Notched12() Overlay {
	return Overlay{overlay(3)}
}

// Notched20 is the overlay for a boss bar split into 20 segments.
func Notched20() Overlay {
	return Overlay{overlay(4)}
}

type overlay uint8

func (o overlay) Uint8() uint8 {
	return uint8(o)
}
