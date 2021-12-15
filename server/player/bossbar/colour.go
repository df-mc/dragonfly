package bossbar

// Colour is the colour of a BossBar.
type Colour struct{ colour }

// Grey is the colour for a grey boss bar.
func Grey() Colour {
	return Colour{colour(0)}
}

// Blue is the colour for a blue boss bar.
func Blue() Colour {
	return Colour{colour(1)}
}

// Red is the colour for a red boss bar.
func Red() Colour {
	return Colour{colour(2)}
}

// Green is the colour for a green boss bar.
func Green() Colour {
	return Colour{colour(3)}
}

// Yellow is the colour for a yellow boss bar.
func Yellow() Colour {
	return Colour{colour(4)}
}

// Purple is the colour for a purple boss bar.
func Purple() Colour {
	return Colour{colour(5)}
}

// White is the colour for a white boss bar.
func White() Colour {
	return Colour{colour(6)}
}

type colour uint8

func (c colour) Uint8() uint8 {
	return uint8(c)
}
