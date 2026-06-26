package block

// SeaGrassType represents a type of seagrass.
type SeaGrassType struct {
	seaGrass
}

type seaGrass uint8

// DefaultSeaGrass is the default, single-block variant of seagrass.
func DefaultSeaGrass() SeaGrassType {
	return SeaGrassType{0}
}

// DoubleTopSeaGrass is the top half of a double-tall seagrass plant.
func DoubleTopSeaGrass() SeaGrassType {
	return SeaGrassType{1}
}

// DoubleBottomSeaGrass is the bottom half of a double-tall seagrass plant.
func DoubleBottomSeaGrass() SeaGrassType {
	return SeaGrassType{2}
}

// SeaGrassTypes returns all valid sea grass types.
func SeaGrassTypes() []SeaGrassType {
	return []SeaGrassType{DefaultSeaGrass(), DoubleTopSeaGrass(), DoubleBottomSeaGrass()}
}

// Uint8 ...
func (s seaGrass) Uint8() uint8 {
	return uint8(s)
}

// String ...
func (s seaGrass) String() string {
	switch s {
	case 0:
		return "default"
	case 1:
		return "double_top"
	case 2:
		return "double_bot"
	}
	panic("unknown seagrass type")
}
