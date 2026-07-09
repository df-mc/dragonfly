package block

// SaplingType represents a sapling species.
type SaplingType struct {
	sapling
}

type sapling uint8

func OakSapling() SaplingType     { return SaplingType{0} }
func SpruceSapling() SaplingType  { return SaplingType{1} }
func BirchSapling() SaplingType   { return SaplingType{2} }
func JungleSapling() SaplingType  { return SaplingType{3} }
func AcaciaSapling() SaplingType  { return SaplingType{4} }
func DarkOakSapling() SaplingType { return SaplingType{5} }
func CherrySapling() SaplingType  { return SaplingType{6} }
func PaleOakSapling() SaplingType { return SaplingType{7} }

func (s sapling) Uint8() uint8 {
	return uint8(s)
}

func (s sapling) Name() string {
	switch s {
	case 0:
		return "Oak Sapling"
	case 1:
		return "Spruce Sapling"
	case 2:
		return "Birch Sapling"
	case 3:
		return "Jungle Sapling"
	case 4:
		return "Acacia Sapling"
	case 5:
		return "Dark Oak Sapling"
	case 6:
		return "Cherry Sapling"
	case 7:
		return "Pale Oak Sapling"
	}
	panic("unknown sapling type")
}

func (s sapling) String() string {
	switch s {
	case 0:
		return "oak_sapling"
	case 1:
		return "spruce_sapling"
	case 2:
		return "birch_sapling"
	case 3:
		return "jungle_sapling"
	case 4:
		return "acacia_sapling"
	case 5:
		return "dark_oak_sapling"
	case 6:
		return "cherry_sapling"
	case 7:
		return "pale_oak_sapling"
	}
	panic("unknown sapling type")
}

// SaplingTypes returns all supported sapling types.
func SaplingTypes() []SaplingType {
	return []SaplingType{
		OakSapling(),
		SpruceSapling(),
		BirchSapling(),
		JungleSapling(),
		AcaciaSapling(),
		DarkOakSapling(),
		CherrySapling(),
		PaleOakSapling(),
	}
}

// Wood returns the wood type grown by the sapling.
func (s SaplingType) Wood() WoodType {
	switch s {
	case OakSapling():
		return OakWood()
	case SpruceSapling():
		return SpruceWood()
	case BirchSapling():
		return BirchWood()
	case JungleSapling():
		return JungleWood()
	case AcaciaSapling():
		return AcaciaWood()
	case DarkOakSapling():
		return DarkOakWood()
	case CherrySapling():
		return CherryWood()
	case PaleOakSapling():
		return PaleOakWood()
	}
	panic("unknown sapling type")
}

// Leaves returns the leaves type grown by the sapling.
func (s SaplingType) Leaves() LeavesType {
	switch s {
	case OakSapling():
		return OakLeaves()
	case SpruceSapling():
		return SpruceLeaves()
	case BirchSapling():
		return BirchLeaves()
	case JungleSapling():
		return JungleLeaves()
	case AcaciaSapling():
		return AcaciaLeaves()
	case DarkOakSapling():
		return DarkOakLeaves()
	case CherrySapling():
		return CherryLeaves()
	case PaleOakSapling():
		return PaleOakLeaves()
	}
	panic("unknown sapling type")
}
