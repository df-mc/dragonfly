package bossbar

import (
	"fmt"
	"strings"
)

// BossBar represents a boss bar that may be sent to a player. It is shown as a purple bar with text above
// it. The health shown by the bar may be changed.
type BossBar struct {
	text   string
	health float64
	c      Colour
}

// New creates a new boss bar with the text passed. The text is formatted according to the rules of
// fmt.Sprintln.
// By default, the boss bar will have a full health bar. To change this, use BossBar.WithHealthPercentage().
// The default colour of the BossBar is Purple. This can be changed using BossBar.WithColour.
func New(text ...any) BossBar {
	return BossBar{text: format(text), health: 1, c: Purple()}
}

// Text returns the text of the boss bar: The text passed when creating the bar using New().
func (bar BossBar) Text() string {
	return bar.text
}

// WithHealthPercentage sets the health percentage of the boss bar. The value passed must be between 0 and 1.
// If a value out of that range is passed, WithHealthPercentage panics.
// The new BossBar with the changed health percentage is returned.
func (bar BossBar) WithHealthPercentage(v float64) BossBar {
	if v < 0 || v > 1 {
		panic("boss bar: value out of range: health percentage must be between 0.0 and 1.0")
	}
	bar.health = v
	return bar
}

// WithColour returns a copy of the BossBar with the Colour passed.
func (bar BossBar) WithColour(c Colour) BossBar {
	bar.c = c
	return bar
}

// HealthPercentage returns the health percentage of the boss bar. The number returned is a value between 0
// and 1, with 0 being an empty boss bar and 1 being a full one.
func (bar BossBar) HealthPercentage() float64 {
	return bar.health
}

// Colour returns the colour of the BossBar.
func (bar BossBar) Colour() Colour {
	return bar.c
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []any) string {
	return strings.TrimSuffix(fmt.Sprintln(a...), "\n")
}
