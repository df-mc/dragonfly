package dialogue

type activation uint8

// ActivationType represents a specific type of activation on a Button. The different types are Click, Close, and Enter.
type ActivationType struct {
	activation
}

func (a activation) Uint8() uint8 {
	return uint8(a)
}

// ActivationClick specifies on activating a button when its clicked.
func ActivationClick() ActivationType {
	return ActivationType{0}
}

// ActivationClose is the close activation.
func ActivationClose() ActivationType {
	return ActivationType{1}
}

// ActivationEnter is the enter activation type.
func ActivationEnter() ActivationType {
	return ActivationType{2}
}

type button uint8

// ButtonType represents the type of Button. URL, COMMAND, and UNKNOWN are all the button types.
type ButtonType struct {
	button
}

func (b button) Uint8() uint8 {
	return uint8(b)
}

// CommandButton is a button meant to execute a command.
func CommandButton() ButtonType {
	return ButtonType{1}
}

// UnknownButton has unknown behaviour as of the moment.
func UnknownButton() ButtonType {
	return ButtonType{2}
}
