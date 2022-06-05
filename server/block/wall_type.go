package block

// WallType represents a type of wall.
type WallType struct {
	wallType
}

// CobblestoneWall returns the cobblestone wall variant.
func CobblestoneWall() WallType {
	return WallType{0}
}

// MossyCobblestoneWall returns the mossy cobblestone wall variant.
func MossyCobblestoneWall() WallType {
	return WallType{1}
}

// GraniteWall returns the granite wall variant.
func GraniteWall() WallType {
	return WallType{2}
}

// DioriteWall returns the diorite wall variant.
func DioriteWall() WallType {
	return WallType{3}
}

// AndesiteWall returns the andesite wall variant.
func AndesiteWall() WallType {
	return WallType{4}
}

// SandstoneWall returns the sandstone wall variant.
func SandstoneWall() WallType {
	return WallType{5}
}

// BrickWall returns the brick wall variant.
func BrickWall() WallType {
	return WallType{6}
}

// StoneBrickWall returns the stone brick wall variant.
func StoneBrickWall() WallType {
	return WallType{7}
}

// MossyStoneBrickWall returns the mossy stone brick wall variant.
func MossyStoneBrickWall() WallType {
	return WallType{8}
}

// NetherBrickWall returns the nether brick wall variant.
func NetherBrickWall() WallType {
	return WallType{9}
}

// EndBrickWall returns the end brick wall variant.
func EndBrickWall() WallType {
	return WallType{10}
}

// PrismarineWall returns the prismarine wall variant.
func PrismarineWall() WallType {
	return WallType{11}
}

// RedSandstoneWall returns the red sandstone wall variant.
func RedSandstoneWall() WallType {
	return WallType{12}
}

// RedNetherBrickWall returns the red nether brick wall variant.
func RedNetherBrickWall() WallType {
	return WallType{13}
}

// BlackstoneWall returns the blackstone wall variant.
func BlackstoneWall() WallType {
	return WallType{14}
}

// PolishedBlackstoneWall returns the polished blackstone wall variant.
func PolishedBlackstoneWall() WallType {
	return WallType{15}
}

// PolishedBlackstoneBrickWall returns the polished blackstone brick wall variant.
func PolishedBlackstoneBrickWall() WallType {
	return WallType{16}
}

// CobbledDeepslateWall returns the cobbled deep slate wall variant.
func CobbledDeepslateWall() WallType {
	return WallType{17}
}

// PolishedDeepslateWall returns the polished deep slate wall variant.
func PolishedDeepslateWall() WallType {
	return WallType{18}
}

// DeepslateBrickWall returns the deep slate brick wall variant.
func DeepslateBrickWall() WallType {
	return WallType{19}
}

// DeepslateTileWall returns the deep slate tile wall variant.
func DeepslateTileWall() WallType {
	return WallType{20}
}

// WallTypes returns a list of all wall types.
func WallTypes() []WallType {
	return []WallType{
		CobblestoneWall(), MossyCobblestoneWall(), GraniteWall(), DioriteWall(), AndesiteWall(), SandstoneWall(), BrickWall(),
		NetherBrickWall(), StoneBrickWall(), MossyStoneBrickWall(), EndBrickWall(), PrismarineWall(), RedSandstoneWall(),
		RedNetherBrickWall(), BlackstoneWall(), PolishedBlackstoneWall(), PolishedBlackstoneBrickWall(), CobbledDeepslateWall(),
		PolishedDeepslateWall(), DeepslateBrickWall(), DeepslateTileWall(),
	}
}

type wallType uint8

// Uint8 returns the wall as a uint8.
func (w wallType) Uint8() uint8 {
	return uint8(w)
}

// String ...
func (w wallType) String() string {
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
	case 14:
		return "blackstone"
	case 15:
		return "polished_blackstone"
	case 16:
		return "polished_blackstone_brick"
	case 17:
		return "cobbled_deepslate"
	case 18:
		return "polished_deepslate"
	case 19:
		return "deepslate_brick"
	case 20:
		return "deepslate_tile"
	}
	panic("unknown wall type")
}

// IsCobblestoneWall returns if the wall is a variant of the cobblestone wall. This is true for every wall that is not
// blackstone or deepslate.
func (w wallType) IsCobblestoneWall() bool {
	return w < 14
}
