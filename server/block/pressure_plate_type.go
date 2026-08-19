package block

// PressurePlateType identifies a pressure plate material.
type PressurePlateType struct {
	pressurePlate
}

type pressurePlate uint8

// StonePressurePlate returns a stone pressure plate type.
func StonePressurePlate() PressurePlateType {
	return PressurePlateType{0}
}

// PolishedBlackstonePressurePlate returns a polished blackstone pressure plate type.
func PolishedBlackstonePressurePlate() PressurePlateType {
	return PressurePlateType{1}
}

// OakPressurePlate returns an oak pressure plate type.
func OakPressurePlate() PressurePlateType {
	return PressurePlateType{2}
}

// SprucePressurePlate returns a spruce pressure plate type.
func SprucePressurePlate() PressurePlateType {
	return PressurePlateType{3}
}

// BirchPressurePlate returns a birch pressure plate type.
func BirchPressurePlate() PressurePlateType {
	return PressurePlateType{4}
}

// JunglePressurePlate returns a jungle pressure plate type.
func JunglePressurePlate() PressurePlateType {
	return PressurePlateType{5}
}

// AcaciaPressurePlate returns an acacia pressure plate type.
func AcaciaPressurePlate() PressurePlateType {
	return PressurePlateType{6}
}

// DarkOakPressurePlate returns a dark oak pressure plate type.
func DarkOakPressurePlate() PressurePlateType {
	return PressurePlateType{7}
}

// MangrovePressurePlate returns a mangrove pressure plate type.
func MangrovePressurePlate() PressurePlateType {
	return PressurePlateType{8}
}

// CherryPressurePlate returns a cherry pressure plate type.
func CherryPressurePlate() PressurePlateType {
	return PressurePlateType{9}
}

// BambooPressurePlate returns a bamboo pressure plate type.
func BambooPressurePlate() PressurePlateType {
	return PressurePlateType{10}
}

// CrimsonPressurePlate returns a crimson pressure plate type.
func CrimsonPressurePlate() PressurePlateType {
	return PressurePlateType{11}
}

// WarpedPressurePlate returns a warped pressure plate type.
func WarpedPressurePlate() PressurePlateType {
	return PressurePlateType{12}
}

// PaleOakPressurePlate returns a pale oak pressure plate type.
func PaleOakPressurePlate() PressurePlateType {
	return PressurePlateType{13}
}

// LightWeightedPressurePlate returns a light weighted pressure plate type.
func LightWeightedPressurePlate() PressurePlateType {
	return PressurePlateType{14}
}

// HeavyWeightedPressurePlate returns a heavy weighted pressure plate type.
func HeavyWeightedPressurePlate() PressurePlateType {
	return PressurePlateType{15}
}

// Uint8 returns the pressure plate type as a uint8.
func (p pressurePlate) Uint8() uint8 {
	return uint8(p)
}

// Wood reports whether the pressure plate behaves like wood.
func (p pressurePlate) Wood() bool {
	return p >= 2 && p <= 13
}

// Flammable reports whether the pressure plate can burn.
func (p pressurePlate) Flammable() bool {
	return p.Wood() && p != CrimsonPressurePlate().pressurePlate && p != WarpedPressurePlate().pressurePlate
}

// Weighted reports whether the plate's signal depends on entity count.
func (p pressurePlate) Weighted() bool {
	return p == 14 || p == 15
}

// Name returns the pressure plate's display name.
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

// String returns the pressure plate's block identifier.
func (p pressurePlate) String() string {
	switch p {
	case 0:
		return "stone_pressure_plate"
	case 1:
		return "polished_blackstone_pressure_plate"
	case 2:
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

// PressurePlateTypes returns all pressure plate types.
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
