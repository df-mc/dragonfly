package block

// PrismarineType represents a type of prismarine.
type PrismarineType struct {
	prismarine
}

type prismarine uint8

// NormalPrismarine is the normal variant of prismarine.
func NormalPrismarine() PrismarineType {
	return PrismarineType{0}
}

// DarkPrismarine is the dark variant of prismarine.
func DarkPrismarine() PrismarineType {
	return PrismarineType{1}
}

// BrickPrismarine is the brick variant of prismarine.
func BrickPrismarine() PrismarineType {
	return PrismarineType{2}
}

// Uint8 returns the prismarine as a uint8.
func (s prismarine) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s prismarine) Name() string {
	switch s {
	case 0:
		return "Prismarine"
	case 1:
		return "Dark Prismarine"
	case 2:
		return "Prismarine Bricks"
	}
	panic("unknown prismarine type")
}

// String ...
func (s prismarine) String() string {
	switch s {
	case 0:
		return "default"
	case 1:
		return "dark"
	case 2:
		return "bricks"
	}
	panic("unknown prismarine type")
}

// PrismarineTypes ...
func PrismarineTypes() []PrismarineType {
	return []PrismarineType{NormalPrismarine(), DarkPrismarine(), BrickPrismarine()}
}
