package dialogue

import "encoding/json"

// Button represents a Button added to Menu. A button contains Name as well as an activation type and a general type.
type Button struct {
	// Name is the name of the button and is displayed to the user.
	Name string
	// Activation is the specific method of activation required by the button. CLICK = 0, CLOSE = 1, ENTER = 2.
	Activation ActivationType
	// Type is the type of button / action it takes.
	Type ButtonType
}

// NewButton returns a new Button with the name, activationType, and buttonType passed.
func NewButton(name string, activationType ActivationType, buttonType ButtonType) Button {
	return Button{Name: name, Activation: activationType, Type: buttonType}
}

// MarshalJSON ...
func (b Button) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"button_name": b.Name,
		"text":        "", // Buttons don't work if this value isn't sent.
		"mode":        b.Activation.Uint8(),
		"type":        b.Type.Uint8(),
	}
	return json.Marshal(data)
}
