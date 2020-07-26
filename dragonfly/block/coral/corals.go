package coral

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/block/colour"
)

// Corals represents a type of coral of a block. Corals, coral fans, and coral blocks carry one of these types.
type Corals struct {
	coral
	Colour colour.Colour
}

// Tube returns the tube coral variant
func Tube() Corals {
	return Corals{coral(0), colour.Blue()}
}

// Brain returns the brain coral variant
func Brain() Corals {
	return Corals{coral(1), colour.Pink()}
}

// Bubble returns the bubble coral variant
func Bubble() Corals {
	return Corals{coral(2), colour.Purple()}
}

// Fire returns the fire coral variant
func Fire() Corals {
	return Corals{coral(3), colour.Red()}
}

// Horn returns the horn coral variant
func Horn() Corals {
	return Corals{coral(4), colour.Yellow()}
}

type coral uint8

// Uint8 returns the coral as a uint8.
func (c coral) Uint8() uint8 {
	return uint8(c)
}

// Name ...
func (c coral) Name() string {
	switch c {
	case 0:
		return "Tube"
	case 1:
		return "Brain"
	case 2:
		return "Bubble"
	case 3:
		return "Fire"
	case 4:
		return "Horn"
	}
	panic("unknown coral type")
}

// FromString ...
func (c coral) FromString(s string) (interface{}, error) {
	switch s {
	case "tube":
		return Corals{coral(0), colour.Blue()}, nil
	case "brain":
		return Corals{coral(1), colour.Pink()}, nil
	case "bubble":
		return Corals{coral(2), colour.Purple()}, nil
	case "fire":
		return Corals{coral(3), colour.Red()}, nil
	case "horn":
		return Corals{coral(4), colour.Yellow()}, nil
	}
	return nil, fmt.Errorf("unexpected coral type '%v', expecting one of 'tube', 'brain', 'bubble', 'fire', or 'horn'", s)
}

// String ...
func (c coral) String() string {
	switch c {
	case 0:
		return "tube"
	case 1:
		return "brain"
	case 2:
		return "bubble"
	case 3:
		return "fire"
	case 4:
		return "horn"
	}
	panic("unknown coral type")
}
