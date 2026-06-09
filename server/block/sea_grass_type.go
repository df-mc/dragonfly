package block

type SeaGrassType struct {
	seaGrass
}

func DefaultSeaGrass() SeaGrassType {
	return SeaGrassType{0}
}

func DoubleTopSeaGrass() SeaGrassType {
	return SeaGrassType{1}
}

func DoubleBottomSeaGrass() SeaGrassType {
	return SeaGrassType{2}
}

func SeaGrassTypes() []SeaGrassType {
	return []SeaGrassType{DefaultSeaGrass(), DoubleTopSeaGrass(), DoubleBottomSeaGrass()}
}

type seaGrass uint8

// Uint8 ...
func (s seaGrass) Uint8() uint8 {
	return uint8(s)
}

func (s seaGrass) Name() string {
	switch s {
	case 0:
		return "Default Sea Grass"
	case 1:
		return "Top Double Sea Grass"
	case 2:
		return "Bottom Double Sea Grass"
	}
	panic("unknown sea grass type")
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
	panic("unknown sea grass type")
}
