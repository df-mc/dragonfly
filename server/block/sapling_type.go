package block

// SaplingType represents a type of sapling. The type of a sapling decides the tree that grows from it.
type SaplingType struct {
	sapling
}

// OakSapling returns oak sapling material.
func OakSapling() SaplingType {
	return SaplingType{0}
}

// SpruceSapling returns spruce sapling material.
func SpruceSapling() SaplingType {
	return SaplingType{1}
}

// BirchSapling returns birch sapling material.
func BirchSapling() SaplingType {
	return SaplingType{2}
}

// JungleSapling returns jungle sapling material.
func JungleSapling() SaplingType {
	return SaplingType{3}
}

// AcaciaSapling returns acacia sapling material.
func AcaciaSapling() SaplingType {
	return SaplingType{4}
}

// DarkOakSapling returns dark oak sapling material.
func DarkOakSapling() SaplingType {
	return SaplingType{5}
}

// CherrySapling returns cherry sapling material.
func CherrySapling() SaplingType {
	return SaplingType{6}
}

// PaleOakSapling returns pale oak sapling material.
func PaleOakSapling() SaplingType {
	return SaplingType{7}
}

// SaplingTypes returns a list of all sapling types.
func SaplingTypes() []SaplingType {
	return []SaplingType{OakSapling(), SpruceSapling(), BirchSapling(), JungleSapling(), AcaciaSapling(),
		DarkOakSapling(), CherrySapling(), PaleOakSapling()}
}

type sapling uint8

// Uint8 returns the sapling as a uint8.
func (s sapling) Uint8() uint8 {
	return uint8(s)
}

// Name ...
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

// String ...
func (s sapling) String() string {
	return s.Wood().String() + "_sapling"
}

// Wood returns the wood of the tree that grows from the sapling.
func (s sapling) Wood() WoodType {
	switch s {
	case 0:
		return OakWood()
	case 1:
		return SpruceWood()
	case 2:
		return BirchWood()
	case 3:
		return JungleWood()
	case 4:
		return AcaciaWood()
	case 5:
		return DarkOakWood()
	case 6:
		return CherryWood()
	case 7:
		return PaleOakWood()
	}
	panic("unknown sapling type")
}

// Leaves returns the leaves of the tree that grows from the sapling.
func (s sapling) Leaves() LeavesType {
	switch s {
	case 0:
		return OakLeaves()
	case 1:
		return SpruceLeaves()
	case 2:
		return BirchLeaves()
	case 3:
		return JungleLeaves()
	case 4:
		return AcaciaLeaves()
	case 5:
		return DarkOakLeaves()
	case 6:
		return CherryLeaves()
	case 7:
		return PaleOakLeaves()
	}
	panic("unknown sapling type")
}
