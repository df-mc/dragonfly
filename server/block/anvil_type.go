package block

// AnvilType represents a type of anvil, such as undamaged, slightly damaged, or very damaged.
type AnvilType struct {
	anvil
}

// UndamagedAnvil returns the undamaged anvil type.
func UndamagedAnvil() AnvilType {
	return AnvilType{0}
}

// SlightlyDamagedAnvil returns the slightly damaged anvil type.
func SlightlyDamagedAnvil() AnvilType {
	return AnvilType{1}
}

// VeryDamagedAnvil returns the very damaged anvil type.
func VeryDamagedAnvil() AnvilType {
	return AnvilType{2}
}

// AnvilTypes returns all anvil types.
func AnvilTypes() []AnvilType {
	return []AnvilType{UndamagedAnvil(), SlightlyDamagedAnvil(), VeryDamagedAnvil()}
}

type anvil uint8

// Uint8 returns the anvil type as a uint8.
func (a anvil) Uint8() uint8 {
	return uint8(a)
}

// String returns the anvil type as a string.
func (a anvil) String() string {
	switch a {
	case 0:
		return "undamaged"
	case 1:
		return "slightly_damaged"
	case 2:
		return "very_damaged"
	}
	panic("should never happen")
}
