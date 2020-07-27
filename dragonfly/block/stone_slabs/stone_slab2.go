package stone_slabs

import "fmt"

// StoneSlab2 represents a type of slab of a block. SmoothStone slabs carry one of these types.
// others carry one of these types.
type StoneSlab2 struct {
	stoneSlab2
}

// RedSandstone returns red sandstone slab material.
func RedSandstone() StoneSlab2 {
	return StoneSlab2{stoneSlab2(0)}
}

// Purpur returns purpur slab material.
func Purpur() StoneSlab2 {
	return StoneSlab2{stoneSlab2(1)}
}

// PrismarineRough returns prismarine slab material.
func PrismarineRough() StoneSlab2 {
	return StoneSlab2{stoneSlab2(2)}
}

// PrismarineDark returns dark prismarine slab material.
func PrismarineDark() StoneSlab2 {
	return StoneSlab2{stoneSlab2(3)}
}

// PrismarineBrick returns prismarine brick slab material.
func PrismarineBrick() StoneSlab2 {
	return StoneSlab2{stoneSlab2(4)}
}

// MossyCobblestone returns mossy cobblestone slab material.
func MossyCobblestone() StoneSlab2 {
	return StoneSlab2{stoneSlab2(5)}
}

// SmoothSandstone returns smooth sandstone slab material.
func SmoothSandstone() StoneSlab2 {
	return StoneSlab2{stoneSlab2(6)}
}

// RedNetherBrick returns red nether brick slab material.
func RedNetherBrick() StoneSlab2 {
	return StoneSlab2{stoneSlab2(7)}
}

type stoneSlab2 uint8

// Uint8 returns the stoneSlab2 as a uint8.
func (w stoneSlab2) Uint8() uint8 {
	return uint8(w)
}

// Name ...
func (w stoneSlab2) Name() string {
	switch w {
	case 0:
		return "Red Sandstone"
	case 1:
		return "Purpur"
	case 2:
		return "Prismarine"
	case 3:
		return "Dark Prismarine"
	case 4:
		return "Prismarine Bricks"
	case 5:
		return "Mossy Cobblestone"
	case 6:
		return "Smooth Sandstone"
	case 7:
		return "Red Nether Brick"
	}
	panic("unknown stone slab type")
}

// FromString ...
func (w stoneSlab2) FromString(s string) (interface{}, error) {
	switch s {
	case "red_sandstone":
		return StoneSlab2{stoneSlab2(0)}, nil
	case "purpur":
		return StoneSlab2{stoneSlab2(1)}, nil
	case "prismarine_rough":
		return StoneSlab2{stoneSlab2(2)}, nil
	case "prismarine_dark":
		return StoneSlab2{stoneSlab2(3)}, nil
	case "prismarine_brick":
		return StoneSlab2{stoneSlab2(4)}, nil
	case "mossy_cobblestone":
		return StoneSlab2{stoneSlab2(5)}, nil
	case "smooth_sandstone":
		return StoneSlab2{stoneSlab2(6)}, nil
	case "red_nether_brick":
		return StoneSlab2{stoneSlab2(7)}, nil
	}
	return nil, fmt.Errorf("unexpected stone slab type '%v', expecting one of 'red_sandstone', 'purpur', 'prismarine_rough', 'prismarine_dark', 'prismarine_brick', 'mossy_cobblestone', 'smooth_sandstone', or 'red_nether_brick'", s)
}

// String ...
func (w stoneSlab2) String() string {
	switch w {
	case 0:
		return "red_sandstone"
	case 1:
		return "purpur"
	case 2:
		return "prismarine_rough"
	case 3:
		return "prismarine_dark"
	case 4:
		return "prismarine_brick"
	case 5:
		return "mossy_cobblestone"
	case 6:
		return "smooth_sandstone"
	case 7:
		return "red_nether_brick"
	}
	panic("unknown stone slab type")
}
