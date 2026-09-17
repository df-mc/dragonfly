package block

// ButtonType identifies a button material.
type ButtonType struct {
	button
}

type button uint8

// StoneButton returns a stone button type.
func StoneButton() ButtonType {
	return ButtonType{0}
}

// PolishedBlackstoneButton returns a polished blackstone button type.
func PolishedBlackstoneButton() ButtonType {
	return ButtonType{1}
}

// OakButton returns an oak button type.
func OakButton() ButtonType {
	return ButtonType{2}
}

// SpruceButton returns a spruce button type.
func SpruceButton() ButtonType {
	return ButtonType{3}
}

// BirchButton returns a birch button type.
func BirchButton() ButtonType {
	return ButtonType{4}
}

// JungleButton returns a jungle button type.
func JungleButton() ButtonType {
	return ButtonType{5}
}

// AcaciaButton returns an acacia button type.
func AcaciaButton() ButtonType {
	return ButtonType{6}
}

// DarkOakButton returns a dark oak button type.
func DarkOakButton() ButtonType {
	return ButtonType{7}
}

// MangroveButton returns a mangrove button type.
func MangroveButton() ButtonType {
	return ButtonType{8}
}

// CherryButton returns a cherry button type.
func CherryButton() ButtonType {
	return ButtonType{9}
}

// BambooButton returns a bamboo button type.
func BambooButton() ButtonType {
	return ButtonType{10}
}

// CrimsonButton returns a crimson button type.
func CrimsonButton() ButtonType {
	return ButtonType{11}
}

// WarpedButton returns a warped button type.
func WarpedButton() ButtonType {
	return ButtonType{12}
}

// PaleOakButton returns a pale oak button type.
func PaleOakButton() ButtonType {
	return ButtonType{13}
}

// Uint8 returns the button type as a uint8.
func (b button) Uint8() uint8 {
	return uint8(b)
}

// Wood reports whether the button behaves like wood.
func (b button) Wood() bool {
	return b >= 2 && b <= 13
}

// Flammable reports whether the button can burn.
func (b button) Flammable() bool {
	return b.Wood() && b != CrimsonButton().button && b != WarpedButton().button
}

// Name returns the button's display name.
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

// String returns the button's block identifier.
func (b button) String() string {
	switch b {
	case 0:
		return "stone_button"
	case 1:
		return "polished_blackstone_button"
	case 2:
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

// ButtonTypes returns all button types.
func ButtonTypes() []ButtonType {
	return []ButtonType{
		StoneButton(),
		PolishedBlackstoneButton(),
		OakButton(),
		SpruceButton(),
		BirchButton(),
		JungleButton(),
		AcaciaButton(),
		DarkOakButton(),
		MangroveButton(),
		CherryButton(),
		BambooButton(),
		CrimsonButton(),
		WarpedButton(),
		PaleOakButton(),
	}
}
