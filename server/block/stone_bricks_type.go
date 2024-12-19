package block

// StoneBricksType represents a type of stone bricks.
type StoneBricksType struct {
	stoneBricks
}

type stoneBricks uint8

// NormalStoneBricks is the normal variant of stone bricks.
func NormalStoneBricks() StoneBricksType {
	return StoneBricksType{0}
}

// MossyStoneBricks is the mossy variant of stone bricks.
func MossyStoneBricks() StoneBricksType {
	return StoneBricksType{1}
}

// CrackedStoneBricks is the cracked variant of stone bricks.
func CrackedStoneBricks() StoneBricksType {
	return StoneBricksType{2}
}

// ChiseledStoneBricks is the chiseled variant of stone bricks.
func ChiseledStoneBricks() StoneBricksType {
	return StoneBricksType{3}
}

// Uint8 returns the stone bricks as a uint8.
func (s stoneBricks) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s stoneBricks) Name() string {
	switch s {
	case 0:
		return "Stone Bricks"
	case 1:
		return "Mossy Stone Bricks"
	case 2:
		return "Cracked Stone Bricks"
	case 3:
		return "Chiseled Stone Bricks"
	}
	panic("unknown stone bricks type")
}

// String ...
func (s stoneBricks) String() string {
	switch s {
	case 0:
		return "stone_bricks"
	case 1:
		return "mossy_stone_bricks"
	case 2:
		return "cracked_stone_bricks"
	case 3:
		return "chiseled_stone_bricks"
	}
	panic("unknown stone bricks type")
}

// StoneBricksTypes ...
func StoneBricksTypes() []StoneBricksType {
	return []StoneBricksType{NormalStoneBricks(), MossyStoneBricks(), CrackedStoneBricks(), ChiseledStoneBricks()}
}
