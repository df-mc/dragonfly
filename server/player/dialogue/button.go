package dialogue

import "encoding/json"

// Button represents a button added to a dialogue menu and consists of just
// text.
type Button struct {
	// Text holds the text displayed on the button. It may use Minecraft
	// formatting codes.
	Text string
}

// MarshalJSON ...
func (b Button) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"button_name": b.Text,
		"text":        "",
		"mode":        0, // "Click" activation
		"type":        1, // "Command" type
	})
}
