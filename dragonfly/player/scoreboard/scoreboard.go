package scoreboard

import (
	"bytes"
	"fmt"
	"strings"
)

// Scoreboard represents a scoreboard that may be sent to a player. The scoreboard is shown on the right side
// of the player's screen.
// Scoreboard implements the io.Writer and io.StringWriter interfaces. fmt.Fprintf and fmt.Fprint may be used
// to write formatted text to the scoreboard.
type Scoreboard struct {
	name string
	p    []byte
}

// New returns a new scoreboard with the display name passed. Once returned, lines may be added to the
// scoreboard to add text to it. The name is formatted according to the rules of fmt.Sprintln.
// Changing the scoreboard after sending it to a player will not update the scoreboard of the player
// automatically: Player.SendScoreboard() must be called again to update it.
func New(name ...interface{}) *Scoreboard {
	return &Scoreboard{name: strings.TrimSuffix(fmt.Sprintln(name...), "\n")}
}

// Name returns the display name of the scoreboard, as passed during the construction of the scoreboard.
func (board *Scoreboard) Name() string {
	return board.name
}

// Write writes a slice of data as text to the scoreboard. Newlines may be written to create a new line on
// the scoreboard.
func (board *Scoreboard) Write(p []byte) (n int, err error) {
	board.p = append(board.p, p...)

	// Scoreboards can have up to 15 lines. (16 including the title.)
	if bytes.Count(board.p, []byte{'\n'}) >= 15 {
		return len(p), fmt.Errorf("write scoreboard: maximum of 15 lines of text exceeded")
	}
	return len(p), nil
}

// WriteString writes a string of text to the scoreboard. Newlines may be written to create a new line on
// the scoreboard.
func (board *Scoreboard) WriteString(s string) (n int, err error) {
	return board.Write([]byte(s))
}

// Bytes returns the data of the Scoreboard as a slice of bytes.
func (board *Scoreboard) Bytes() []byte {
	return board.p
}
