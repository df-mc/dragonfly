package stone_slabs

import "fmt"

// StoneSlab4 represents a type of slab of a block. SmoothStone slabs carry one of these types.
// others carry one of these types.
type StoneSlab4 struct {
	stoneSlab4
}

// MossyStoneBrick returns mossy stone slab material.
func MossyStoneBrick() StoneSlab4 {
	return StoneSlab4{stoneSlab4(0)}
}

// SmoothQuartz returns smooth quartz slab material.
func SmoothQuartz() StoneSlab4 {
	return StoneSlab4{stoneSlab4(1)}
}

// Stone returns stone slab material.
func Stone() StoneSlab4 {
	return StoneSlab4{stoneSlab4(2)}
}

// CutSandstone returns cut sandstone slab material.
func CutSandstone() StoneSlab4 {
	return StoneSlab4{stoneSlab4(3)}
}

// CutRedSandstone returns cut red sandstone slab material.
func CutRedSandstone() StoneSlab4 {
	return StoneSlab4{stoneSlab4(4)}
}

type stoneSlab4 uint8

// Uint8 returns the stoneSlab4 as a uint8.
func (w stoneSlab4) Uint8() uint8 {
	return uint8(w)
}

// Name ...
func (w stoneSlab4) Name() string {
	switch w {
	case 0:
		return "Mossy Stone Brick"
	case 1:
		return "Smooth Quartz"
	case 2:
		return "Stone"
	case 3:
		return "Cut Sandstone"
	case 4:
		return "Cut Red Sandstone"
	}
	panic("unknown stone slab type")
}

// FromString ...
func (w stoneSlab4) FromString(s string) (interface{}, error) {
	switch s {
	case "mossy_stone_brick":
		return StoneSlab4{stoneSlab4(0)}, nil
	case "smooth_quartz":
		return StoneSlab4{stoneSlab4(1)}, nil
	case "stone":
		return StoneSlab4{stoneSlab4(2)}, nil
	case "cut_sandstone":
		return StoneSlab4{stoneSlab4(3)}, nil
	case "cut_red_sandstone":
		return StoneSlab4{stoneSlab4(4)}, nil
	}
	return nil, fmt.Errorf("unexpected stone slab type '%v', expecting one of 'mossy_stone_brick', 'smooth_quartz', 'stone', 'cut_sandstone', or 'cut_red_sandstone'", s)
}

// String ...
func (w stoneSlab4) String() string {
	switch w {
	case 0:
		return "mossy_stone_brick"
	case 1:
		return "smooth_quartz"
	case 2:
		return "stone"
	case 3:
		return "cut_sandstone"
	case 4:
		return "cut_red_sandstone"
	}
	panic("unknown stone slab type")
}
