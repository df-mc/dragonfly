package block

// ShulkerBoxType represents a type of shulker box.
type ShulkerBoxType struct {
	shulkerBox
}

type shulkerBox uint8

// NormalShulkerBox is the normal variant of the shulker box.
func NormalShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{0}
}

// WhiteShulkerBox is the white variant of the shulker box.
func WhiteShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{1}
}

// OrangeShulkerBox is the orange variant of the shulker box.
func OrangeShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{2}
}

// MagentaShulkerBox is the magenta variant of the shulker box.
func MagentaShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{3}
}

// LightBlueShulkerBox is the light blue variant of the shulker box.
func LightBlueShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{4}
}

// YellowShulkerBox is the yellow variant of the shulker box.
func YellowShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{5}
}

// LimeShulkerBox is the lime variant of the shulker box.
func LimeShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{6}
}

// PinkShulkerBox is the pink variant of the shulker box.
func PinkShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{7}
}

// GrayShulkerBox is the gray variant of the shulker box.
func GrayShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{8}
}

// LightGrayShulkerBox is the light gray variant of the shulker box.
func LightGrayShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{9}
}

// CyanShulkerBox is the cyan variant of the shulker box.
func CyanShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{10}
}

// PurpleShulkerBox is the purple variant of the shulker box.
func PurpleShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{11}
}

// BlueShulkerBox is the blue variant of the shulker box.
func BlueShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{12}
}

// BrownShulkerBox is the brown variant of the shulker box.
func BrownShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{13}
}

// GreenShulkerBox is the green variant of the shulker box.
func GreenShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{14}
}

// RedShulkerBox is the red variant of the shulker box.
func RedShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{15}
}

// BlackShulkerBox is the black variant of the shulker box.
func BlackShulkerBox() ShulkerBoxType {
	return ShulkerBoxType{16}
}

// Uint8 returns the shulker box type as a uint8.
func (s shulkerBox) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s shulkerBox) Name() string {
	switch s {
	case 0:
		return "Shulker Box"
	case 1:
		return "White Shulker Box"
	case 2:
		return "Orange Shulker Box"
	case 3:
		return "Magenta Shulker Box"
	case 4:
		return "Light Blue Shulker Box"
	case 5:
		return "Yellow Shulker Box"
	case 6:
		return "Lime Shulker Box"
	case 7:
		return "Pink Shulker Box"
	case 8:
		return "Gray Shulker Box"
	case 9:
		return "Light Gray Shulker Box"
	case 10:
		return "Cyan Shulker Box"
	case 11:
		return "Purple Shulker Box"
	case 12:
		return "Blue Shulker Box"
	case 13:
		return "Brown Shulker Box"
	case 14:
		return "Green Shulker Box"
	case 15:
		return "Red Shulker Box"
	case 16:
		return "Black Shulker Box"
	}

	panic("unknown shulker box type")
}

// String ...
func (s shulkerBox) String() string {
	switch s {
	case 0:
		return "undyed_shulker_box"
	case 1:
		return "white_shulker_box"
	case 2:
		return "orange_shulker_box"
	case 3:
		return "magenta_shulker_box"
	case 4:
		return "light_blue_shulker_box"
	case 5:
		return "yellow_shulker_box"
	case 6:
		return "lime_shulker_box"
	case 7:
		return "pink_shulker_box"
	case 8:
		return "gray_shulker_box"
	case 9:
		return "light_gray_shulker_box"
	case 10:
		return "cyan_shulker_box"
	case 11:
		return "purple_shulker_box"
	case 12:
		return "blue_shulker_box"
	case 13:
		return "brown_shulker_box"
	case 14:
		return "green_shulker_box"
	case 15:
		return "red_shulker_box"
	case 16:
		return "black_shulker_box"
	}

	panic("unkown shulker box type")
}

// ShulkerBoxTypes returns all shulker box types.
func ShulkerBoxTypes() []ShulkerBoxType {
	return []ShulkerBoxType{
		NormalShulkerBox(),
		WhiteShulkerBox(),
		OrangeShulkerBox(),
		MagentaShulkerBox(),
		LightBlueShulkerBox(),
		YellowShulkerBox(),
		LimeShulkerBox(),
		PinkShulkerBox(),
		GrayShulkerBox(),
		LightGrayShulkerBox(),
		CyanShulkerBox(),
		PurpleShulkerBox(),
		BlueShulkerBox(),
		BrownShulkerBox(),
		GreenShulkerBox(),
		RedShulkerBox(),
		BlackShulkerBox(),
	}
}
