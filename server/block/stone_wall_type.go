package block

// StoneWallType represents a type of cobblestone wall.
type StoneWallType struct {
	stoneWallType
}

// CobblestoneWall returns the cobblestone wall variant.
func CobblestoneWall() StoneWallType {
	return StoneWallType{0}
}

// MossyCobblestoneWall returns the mossy cobblestone wall variant.
func MossyCobblestoneWall() StoneWallType {
	return StoneWallType{1}
}

// GraniteWall returns the granite wall variant.
func GraniteWall() StoneWallType {
	return StoneWallType{2}
}

// DioriteWall returns the diorite wall variant.
func DioriteWall() StoneWallType {
	return StoneWallType{3}
}

// AndesiteWall returns the andesite wall variant.
func AndesiteWall() StoneWallType {
	return StoneWallType{4}
}

// SandstoneWall returns the sandstone wall variant.
func SandstoneWall() StoneWallType {
	return StoneWallType{5}
}

// BrickWall returns the brick wall variant.
func BrickWall() StoneWallType {
	return StoneWallType{6}
}

// StoneBrickWall returns the stone brick wall variant.
func StoneBrickWall() StoneWallType {
	return StoneWallType{7}
}

// MossyStoneBrickWall returns the mossy stone brick wall variant.
func MossyStoneBrickWall() StoneWallType {
	return StoneWallType{8}
}

// NetherBrickWall returns the nether brick wall variant.
func NetherBrickWall() StoneWallType {
	return StoneWallType{9}
}

// EndBrickWall returns the end brick wall variant.
func EndBrickWall() StoneWallType {
	return StoneWallType{10}
}

// PrismarineWall returns the prismarine wall variant.
func PrismarineWall() StoneWallType {
	return StoneWallType{11}
}

// RedSandstoneWall returns the red sandstone wall variant.
func RedSandstoneWall() StoneWallType {
	return StoneWallType{12}
}

// RedNetherBrickWall returns the red nether brick wall variant.
func RedNetherBrickWall() StoneWallType {
	return StoneWallType{13}
}

// StoneWallTypes returns a list of all wall types.
func StoneWallTypes() []StoneWallType {
	return []StoneWallType{
		CobblestoneWall(), MossyCobblestoneWall(), GraniteWall(), DioriteWall(), AndesiteWall(), SandstoneWall(), BrickWall(),
		NetherBrickWall(), StoneBrickWall(), MossyStoneBrickWall(), EndBrickWall(), PrismarineWall(), RedSandstoneWall(),
		RedNetherBrickWall(),
	}
}

type stoneWallType uint8

// Uint8 returns the cobblestone wall as a uint8.
func (w stoneWallType) Uint8() uint8 {
	return uint8(w)
}

// String ...
func (w stoneWallType) String() string {
	switch w {
	case 0:
		return "cobblestone"
	case 1:
		return "mossy_cobblestone"
	case 2:
		return "granite"
	case 3:
		return "diorite"
	case 4:
		return "andesite"
	case 5:
		return "sandstone"
	case 6:
		return "brick"
	case 7:
		return "stone_brick"
	case 8:
		return "mossy_stone_brick"
	case 9:
		return "nether_brick"
	case 10:
		return "end_brick"
	case 11:
		return "prismarine"
	case 12:
		return "red_sandstone"
	case 13:
		return "red_nether_brick"
	}
	panic("unknown cobblestone wall type")
}
