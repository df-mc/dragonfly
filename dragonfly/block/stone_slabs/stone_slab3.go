package stone_slabs

import "fmt"

// StoneSlab3 represents a type of slab of a block. SmoothStone slabs carry one of these types.
// others carry one of these types.
type StoneSlab3 struct {
	stoneSlab3
}

// SmoothStone returns smooth stone slab material.
func EndStoneBrick() StoneSlab3 {
	return StoneSlab3{stoneSlab3(0)}
}

// SmoothRedSandstone returns smooth red sandstone slab material.
func SmoothRedSandstone() StoneSlab3 {
	return StoneSlab3{stoneSlab3(1)}
}

// PolishedAndesite returns polished andesite slab material.
func PolishedAndesite() StoneSlab3 {
	return StoneSlab3{stoneSlab3(2)}
}

// Andesite returns andesite slab material.
func Andesite() StoneSlab3 {
	return StoneSlab3{stoneSlab3(3)}
}

// Diorite returns diorite slab material.
func Diorite() StoneSlab3 {
	return StoneSlab3{stoneSlab3(4)}
}

// PolishedDiorite returns polished diorite slab material.
func PolishedDiorite() StoneSlab3 {
	return StoneSlab3{stoneSlab3(5)}
}

// Granite returns granite slab material.
func Granite() StoneSlab3 {
	return StoneSlab3{stoneSlab3(6)}
}

// PolishedGranite returns polished granite slab material.
func PolishedGranite() StoneSlab3 {
	return StoneSlab3{stoneSlab3(7)}
}

type stoneSlab3 uint8

// Uint8 returns the stoneSlab3 as a uint8.
func (w stoneSlab3) Uint8() uint8 {
	return uint8(w)
}

// Name ...
func (w stoneSlab3) Name() string {
	switch w {
	case 0:
		return "End Stone Brick"
	case 1:
		return "Smooth Red Sandstone"
	case 2:
		return "Polished Andesite"
	case 3:
		return "Andesite"
	case 4:
		return "Diorite"
	case 5:
		return "Polished Diorite"
	case 6:
		return "Granite"
	case 7:
		return "Polished Granite"
	}
	panic("unknown stone slab type")
}

// FromString ...
func (w stoneSlab3) FromString(s string) (interface{}, error) {
	switch s {
	case "end_stone_brick":
		return StoneSlab3{stoneSlab3(0)}, nil
	case "smooth_red_sandstone":
		return StoneSlab3{stoneSlab3(1)}, nil
	case "polished_andesite":
		return StoneSlab3{stoneSlab3(2)}, nil
	case "andesite":
		return StoneSlab3{stoneSlab3(3)}, nil
	case "diorite":
		return StoneSlab3{stoneSlab3(4)}, nil
	case "polished_diorite":
		return StoneSlab3{stoneSlab3(5)}, nil
	case "granite":
		return StoneSlab3{stoneSlab3(6)}, nil
	case "polished_granite":
		return StoneSlab3{stoneSlab3(7)}, nil
	}
	return nil, fmt.Errorf("unexpected stone slab type '%v', expecting one of 'end_stone_brick', 'smooth_red_sandstone', 'polished_andesite', 'andesite', 'diorite', 'polished_diorite', 'granite', or 'polished_granite'", s)
}

// String ...
func (w stoneSlab3) String() string {
	switch w {
	case 0:
		return "end_stone_brick"
	case 1:
		return "smooth_red_sandstone"
	case 2:
		return "polished_andesite"
	case 3:
		return "andesite"
	case 4:
		return "diorite"
	case 5:
		return "polished_diorite"
	case 6:
		return "granite"
	case 7:
		return "polished_granite"
	}
	panic("unknown stone slab type")
}
