package block

// ButtonType represents the material a button is made of.
type ButtonType struct {
	button
}

type button uint8

// StoneButton returns the stone button variant.
func StoneButton() ButtonType {
	return ButtonType{0}
}

// PolishedBlackstoneButton returns the polished blackstone button variant.
func PolishedBlackstoneButton() ButtonType {
	return ButtonType{1}
}

// OakButton returns the oak button variant.
func OakButton() ButtonType {
	return ButtonType{2}
}

// SpruceButton returns the spruce button variant.
func SpruceButton() ButtonType {
	return ButtonType{3}
}

// BirchButton returns the birch button variant.
func BirchButton() ButtonType {
	return ButtonType{4}
}

// JungleButton returns the jungle button variant.
func JungleButton() ButtonType {
	return ButtonType{5}
}

// AcaciaButton returns the acacia button variant.
func AcaciaButton() ButtonType {
	return ButtonType{6}
}

// DarkOakButton returns the dark oak button variant.
func DarkOakButton() ButtonType {
	return ButtonType{7}
}

// MangroveButton returns the mangrove button variant.
func MangroveButton() ButtonType {
	return ButtonType{8}
}

// CherryButton returns the cherry button variant.
func CherryButton() ButtonType {
	return ButtonType{9}
}

// BambooButton returns the bamboo button variant.
func BambooButton() ButtonType {
	return ButtonType{10}
}

// CrimsonButton returns the crimson button variant.
func CrimsonButton() ButtonType {
	return ButtonType{11}
}

// WarpedButton returns the warped button variant.
func WarpedButton() ButtonType {
	return ButtonType{12}
}

// PaleOakButton returns the pale oak button variant.
func PaleOakButton() ButtonType {
	return ButtonType{13}
}

// Uint8 returns the button type as a uint8.
func (b button) Uint8() uint8 {
	return uint8(b)
}

// Wood reports whether the button is made of wood, giving it a longer press
// duration and making it usable as furnace fuel.
func (b button) Wood() bool {
	return b >= 2
}

// Name ...
func (b button) Name() string {
	switch b {
	case 0:
		return "Stone Button"
	case 1:
		return "Polished Blackstone Button"
	case 2:
		return "Oak Button"
	case 3:
		return "Spruce Button"
	case 4:
		return "Birch Button"
	case 5:
		return "Jungle Button"
	case 6:
		return "Acacia Button"
	case 7:
		return "Dark Oak Button"
	case 8:
		return "Mangrove Button"
	case 9:
		return "Cherry Button"
	case 10:
		return "Bamboo Button"
	case 11:
		return "Crimson Button"
	case 12:
		return "Warped Button"
	case 13:
		return "Pale Oak Button"
	}
	panic("unknown button type")
}

// String ...
func (b button) String() string {
	switch b {
	case 0:
		return "stone_button"
	case 1:
		return "polished_blackstone_button"
	case 2:
		// Oak buttons use the legacy wooden identifier.
		return "wooden_button"
	case 3:
		return "spruce_button"
	case 4:
		return "birch_button"
	case 5:
		return "jungle_button"
	case 6:
		return "acacia_button"
	case 7:
		return "dark_oak_button"
	case 8:
		return "mangrove_button"
	case 9:
		return "cherry_button"
	case 10:
		return "bamboo_button"
	case 11:
		return "crimson_button"
	case 12:
		return "warped_button"
	case 13:
		return "pale_oak_button"
	}
	panic("unknown button type")
}

// ButtonTypes ...
func ButtonTypes() []ButtonType {
	types := make([]ButtonType, 14)
	for i := range types {
		types[i] = ButtonType{button(i)}
	}
	return types
}
