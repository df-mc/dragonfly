package block

// SandstoneType represents a type of sandstone.
type SandstoneType struct {
	sandstone
}

type sandstone uint8

// NormalSandstone is the normal variant of sandstone.
func NormalSandstone() SandstoneType {
	return SandstoneType{0}
}

// CutSandstone is the cut variant of sandstone.
func CutSandstone() SandstoneType {
	return SandstoneType{1}
}

// ChiseledSandstone is the chiseled variant of sandstone.
func ChiseledSandstone() SandstoneType {
	return SandstoneType{2}
}

// SmoothSandstone is the smooth variant of sandstone.
func SmoothSandstone() SandstoneType {
	return SandstoneType{3}
}

// Uint8 returns the sandstone as a uint8.
func (s sandstone) Uint8() uint8 {
	return uint8(s)
}

// Hardness ...
func (s sandstone) Hardness() float64 {
	switch s {
	case 3:
		return 2.0
	}
	return 0.8
}

// Name ...
func (s sandstone) Name() string {
	switch s {
	case 0:
		return "Sandstone"
	case 1:
		return "Cut Sandstone"
	case 2:
		return "Chiseled Sandstone"
	case 3:
		return "Smooth Sandstone"
	}
	panic("unknown sandstone type")
}

// String ...
func (s sandstone) String() string {
	switch s {
	case 0:
		return "default"
	case 1:
		return "cut"
	case 2:
		return "heiroglyphs"
	case 3:
		return "smooth"
	}
	panic("unknown sandstone type")
}

// SlabAble checks if a slab with this type can exist.
func (s sandstone) SlabAble() bool {
	return s != ChiseledSandstone().sandstone
}

// StairAble checks if a stair with this type can exist.
func (s sandstone) StairAble() bool {
	return s == NormalSandstone().sandstone || s == SmoothSandstone().sandstone
}

// SandstoneTypes ...
func SandstoneTypes() []SandstoneType {
	return []SandstoneType{NormalSandstone(), CutSandstone(), ChiseledSandstone(), SmoothSandstone()}
}
