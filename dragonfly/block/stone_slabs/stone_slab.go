package stone_slabs

import "fmt"

// StoneSlab represents a type of slab of a block. SmoothStone slabs carry one of these types.
// others carry one of these types.
type StoneSlab struct {
	stoneSlab
}

// SmoothStone returns smooth stone slab material.
func SmoothStone() StoneSlab {
	return StoneSlab{stoneSlab(0)}
}

// Sandstone returns sandstone slab material.
func Sandstone() StoneSlab {
	return StoneSlab{stoneSlab(1)}
}

// Wood returns wooden slab material.
func Wood() StoneSlab {
	return StoneSlab{stoneSlab(2)}
}

// Cobblestone returns cobblestone slab material.
func Cobblestone() StoneSlab {
	return StoneSlab{stoneSlab(3)}
}

// Bricks returns bricks slab material.
func Bricks() StoneSlab {
	return StoneSlab{stoneSlab(4)}
}

// StoneBrick returns stone brick slab material.
func StoneBrick() StoneSlab {
	return StoneSlab{stoneSlab(5)}
}

// Quartz returns quartz slab material.
func Quartz() StoneSlab {
	return StoneSlab{stoneSlab(6)}
}

// NetherBrick returns nether brick slab material.
func NetherBrick() StoneSlab {
	return StoneSlab{stoneSlab(7)}
}

type stoneSlab uint8

// Uint8 returns the stoneSlab as a uint8.
func (w stoneSlab) Uint8() uint8 {
	return uint8(w)
}

// Name ...
func (w stoneSlab) Name() string {
	switch w {
	case 0:
		return "Smooth Stone"
	case 1:
		return "Sandstone"
	case 2:
		return "Wooden"
	case 3:
		return "Cobblestone"
	case 4:
		return "Bricks"
	case 5:
		return "Stone Bricks"
	case 6:
		return "Quartz"
	case 7:
		return "Nether Brick"
	}
	panic("unknown stone slab type")
}

// FromString ...
func (w stoneSlab) FromString(s string) (interface{}, error) {
	switch s {
	case "smooth_stone":
		return StoneSlab{stoneSlab(0)}, nil
	case "sandstone":
		return StoneSlab{stoneSlab(1)}, nil
	case "wood":
		return StoneSlab{stoneSlab(2)}, nil
	case "cobblestone":
		return StoneSlab{stoneSlab(3)}, nil
	case "brick":
		return StoneSlab{stoneSlab(4)}, nil
	case "stone_brick":
		return StoneSlab{stoneSlab(5)}, nil
	case "quartz":
		return StoneSlab{stoneSlab(6)}, nil
	case "nether_brick":
		return StoneSlab{stoneSlab(7)}, nil
	}
	return nil, fmt.Errorf("unexpected stone slab type '%v', expecting one of 'smooth_stone', 'sandstone', 'wood', 'cobblestone', 'brick', 'stone_brick', 'quartz', or 'nether_brick'", s)
}

// String ...
func (w stoneSlab) String() string {
	switch w {
	case 0:
		return "smooth_stone"
	case 1:
		return "sandstone"
	case 2:
		return "wood"
	case 3:
		return "cobblestone"
	case 4:
		return "brick"
	case 5:
		return "stone_brick"
	case 6:
		return "quartz"
	case 7:
		return "nether_brick"
	}
	panic("unknown stone slab type")
}
