package block

// WallType represents a type of wall.
type WallType struct {
	wallType
}

// BlackstoneWall returns the blackstone wall variant.
func BlackstoneWall() WallType {
	return WallType{0}
}

// PolishedBlackstoneWall returns the polished blackstone wall variant.
func PolishedBlackstoneWall() WallType {
	return WallType{1}
}

// PolishedBlackstoneBrickWall returns the polished blackstone brick wall variant.
func PolishedBlackstoneBrickWall() WallType {
	return WallType{2}
}

// CobbledDeepslateWall returns the cobbled deep slate wall variant.
func CobbledDeepslateWall() WallType {
	return WallType{3}
}

// PolishedDeepslateWall returns the polished deep slate wall variant.
func PolishedDeepslateWall() WallType {
	return WallType{4}
}

// DeepslateBrickWall returns the deep slate brick wall variant.
func DeepslateBrickWall() WallType {
	return WallType{5}
}

// DeepslateTileWall returns the deep slate tile wall variant.
func DeepslateTileWall() WallType {
	return WallType{6}
}

// WallTypes returns a list of all wall types.
func WallTypes() []WallType {
	return []WallType{
		BlackstoneWall(), PolishedBlackstoneWall(), PolishedBlackstoneBrickWall(), CobbledDeepslateWall(),
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
		return "blackstone"
	case 1:
		return "polished_blackstone"
	case 2:
		return "polished_blackstone_brick"
	case 3:
		return "cobbled_deepslate"
	case 4:
		return "polished_deepslate"
	case 5:
		return "deepslate_brick"
	case 6:
		return "deepslate_tile"
	}
	panic("unknown wall type")
}
