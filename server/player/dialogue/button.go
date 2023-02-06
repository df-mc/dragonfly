package dialogue

import "encoding/json"

// Button represents a Button added to Menu. A button contains Name as well as an activation type and a general type.
type Button struct {
	// Name is the name of the button and is displayed to the user.
	Name string
	// Activation is the specific method of activation required by the button.
	Activation ActivationType
}

// NewButton returns a new Button with the name, activationType, and buttonType passed.
func NewButton(name string, activationType ActivationType) Button {
	return Button{Name: name, Activation: activationType}
}

// MarshalJSON ...
func (b Button) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"button_name": b.Name,
		"text":        "", // Buttons don't work if this value isn't sent.
		"mode":        b.Activation.Uint8(),
		"type":        1,
	}
	return json.Marshal(data)
}
