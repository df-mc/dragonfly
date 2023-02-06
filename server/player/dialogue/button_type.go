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
