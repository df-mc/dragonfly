package block

// PressurePlateType represents the material a pressure plate is made of,
// including the weighted gold and iron variants.
type PressurePlateType struct {
	pressurePlate
}

type pressurePlate uint8

// StonePressurePlate returns the stone pressure plate variant.
func StonePressurePlate() PressurePlateType {
	return PressurePlateType{0}
}

// PolishedBlackstonePressurePlate returns the polished blackstone pressure plate variant.
func PolishedBlackstonePressurePlate() PressurePlateType {
	return PressurePlateType{1}
}

// OakPressurePlate returns the oak pressure plate variant.
func OakPressurePlate() PressurePlateType {
	return PressurePlateType{2}
}

// SprucePressurePlate returns the spruce pressure plate variant.
func SprucePressurePlate() PressurePlateType {
	return PressurePlateType{3}
}

// BirchPressurePlate returns the birch pressure plate variant.
func BirchPressurePlate() PressurePlateType {
	return PressurePlateType{4}
}

// JunglePressurePlate returns the jungle pressure plate variant.
func JunglePressurePlate() PressurePlateType {
	return PressurePlateType{5}
}

// AcaciaPressurePlate returns the acacia pressure plate variant.
func AcaciaPressurePlate() PressurePlateType {
	return PressurePlateType{6}
}

// DarkOakPressurePlate returns the dark oak pressure plate variant.
func DarkOakPressurePlate() PressurePlateType {
	return PressurePlateType{7}
}

// MangrovePressurePlate returns the mangrove pressure plate variant.
func MangrovePressurePlate() PressurePlateType {
	return PressurePlateType{8}
}

// CherryPressurePlate returns the cherry pressure plate variant.
func CherryPressurePlate() PressurePlateType {
	return PressurePlateType{9}
}

// BambooPressurePlate returns the bamboo pressure plate variant.
func BambooPressurePlate() PressurePlateType {
	return PressurePlateType{10}
}

// CrimsonPressurePlate returns the crimson pressure plate variant.
func CrimsonPressurePlate() PressurePlateType {
	return PressurePlateType{11}
}

// WarpedPressurePlate returns the warped pressure plate variant.
func WarpedPressurePlate() PressurePlateType {
	return PressurePlateType{12}
}

// PaleOakPressurePlate returns the pale oak pressure plate variant.
func PaleOakPressurePlate() PressurePlateType {
	return PressurePlateType{13}
}

// LightWeightedPressurePlate returns the light weighted (gold) pressure plate
// variant, which emits one power level per entity on it.
func LightWeightedPressurePlate() PressurePlateType {
	return PressurePlateType{14}
}

// HeavyWeightedPressurePlate returns the heavy weighted (iron) pressure plate
// variant, which emits one power level per ten entities on it.
func HeavyWeightedPressurePlate() PressurePlateType {
	return PressurePlateType{15}
}

// Uint8 returns the pressure plate type as a uint8.
func (p pressurePlate) Uint8() uint8 {
	return uint8(p)
}

// Wood reports whether the pressure plate is made of wood, making it react to
// every entity.
func (p pressurePlate) Wood() bool {
	return p >= 2 && p <= 13
}

// Flammable reports whether the pressure plate can burn, making it usable as
// furnace fuel. Crimson and warped plates are wooden but do not burn.
func (p pressurePlate) Flammable() bool {
	return p.Wood() && p != CrimsonPressurePlate().pressurePlate && p != WarpedPressurePlate().pressurePlate
}

// Weighted reports whether the pressure plate emits an analog power level
// based on the number of entities on it.
func (p pressurePlate) Weighted() bool {
	return p == 14 || p == 15
}

// Name ...
func (p pressurePlate) Name() string {
	switch p {
	case 0:
		return "Stone Pressure Plate"
	case 1:
		return "Polished Blackstone Pressure Plate"
	case 2:
		return "Oak Pressure Plate"
	case 3:
		return "Spruce Pressure Plate"
	case 4:
		return "Birch Pressure Plate"
	case 5:
		return "Jungle Pressure Plate"
	case 6:
		return "Acacia Pressure Plate"
	case 7:
		return "Dark Oak Pressure Plate"
	case 8:
		return "Mangrove Pressure Plate"
	case 9:
		return "Cherry Pressure Plate"
	case 10:
		return "Bamboo Pressure Plate"
	case 11:
		return "Crimson Pressure Plate"
	case 12:
		return "Warped Pressure Plate"
	case 13:
		return "Pale Oak Pressure Plate"
	case 14:
		return "Light Weighted Pressure Plate"
	case 15:
		return "Heavy Weighted Pressure Plate"
	}
	panic("unknown pressure plate type")
}

// String ...
func (p pressurePlate) String() string {
	switch p {
	case 0:
		return "stone_pressure_plate"
	case 1:
		return "polished_blackstone_pressure_plate"
	case 2:
		// Oak pressure plates use the legacy wooden identifier.
		return "wooden_pressure_plate"
	case 3:
		return "spruce_pressure_plate"
	case 4:
		return "birch_pressure_plate"
	case 5:
		return "jungle_pressure_plate"
	case 6:
		return "acacia_pressure_plate"
	case 7:
		return "dark_oak_pressure_plate"
	case 8:
		return "mangrove_pressure_plate"
	case 9:
		return "cherry_pressure_plate"
	case 10:
		return "bamboo_pressure_plate"
	case 11:
		return "crimson_pressure_plate"
	case 12:
		return "warped_pressure_plate"
	case 13:
		return "pale_oak_pressure_plate"
	case 14:
		return "light_weighted_pressure_plate"
	case 15:
		return "heavy_weighted_pressure_plate"
	}
	panic("unknown pressure plate type")
}

// PressurePlateTypes ...
func PressurePlateTypes() []PressurePlateType {
	return []PressurePlateType{
		StonePressurePlate(),
		PolishedBlackstonePressurePlate(),
		OakPressurePlate(),
		SprucePressurePlate(),
		BirchPressurePlate(),
		JunglePressurePlate(),
		AcaciaPressurePlate(),
		DarkOakPressurePlate(),
		MangrovePressurePlate(),
		CherryPressurePlate(),
		BambooPressurePlate(),
		CrimsonPressurePlate(),
		WarpedPressurePlate(),
		PaleOakPressurePlate(),
		LightWeightedPressurePlate(),
		HeavyWeightedPressurePlate(),
	}
}
