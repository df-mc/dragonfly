package block

// ButtonType represents a type of button.
type ButtonType struct {
	button

	// Wood is the type of wood of the button.
	wood WoodType
}

type button uint8

// WoodButton returns the wood button type.
func WoodButton(w WoodType) ButtonType {
	return ButtonType{0, w}
}

// StoneButton returns the stone button type.
func StoneButton() ButtonType {
	return ButtonType{button: 1}
}

// PolishedBlackstoneButton returns the polished blackstone button type.
func PolishedBlackstoneButton() ButtonType {
	return ButtonType{button: 2}
}

// Uint8 ...
func (b ButtonType) Uint8() uint8 {
	return b.wood.Uint8() | uint8(b.button)<<4
}

// Name ...
func (b ButtonType) Name() string {
	switch b.button {
	case 0:
		return b.wood.Name() + " Button"
	case 1:
		return "Stone Button"
	case 2:
		return "Polished Blackstone Button"
	}
	panic("unknown button type")
}

// String ...
func (b ButtonType) String() string {
	switch b.button {
	case 0:
		if b.wood == OakWood() {
			return "wooden"
		}
		return b.wood.String()
	case 1:
		return "stone"
	case 2:
		return "polished_blackstone"
	}
	panic("unknown button type")
}

// ButtonTypes ...
func ButtonTypes() []ButtonType {
	types := []ButtonType{StoneButton(), PolishedBlackstoneButton()}
	for _, w := range WoodTypes() {
		types = append(types, WoodButton(w))
	}
	return types
}
