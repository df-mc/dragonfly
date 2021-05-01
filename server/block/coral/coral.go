package coral

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/colour"
)

// Coral represents a type of coral of a block. Coral, coral fans, and coral blocks carry one of these types.
type Coral struct {
	coral
	Colour colour.Colour
}

// Tube returns the tube coral variant
func Tube() Coral {
	return Coral{coral(0), colour.Blue()}
}

// Brain returns the brain coral variant
func Brain() Coral {
	return Coral{coral(1), colour.Pink()}
}

// Bubble returns the bubble coral variant
func Bubble() Coral {
	return Coral{coral(2), colour.Purple()}
}

// Fire returns the fire coral variant
func Fire() Coral {
	return Coral{coral(3), colour.Red()}
}

// Horn returns the horn coral variant
func Horn() Coral {
	return Coral{coral(4), colour.Yellow()}
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
		return Coral{coral(0), colour.Blue()}, nil
	case "brain":
		return Coral{coral(1), colour.Pink()}, nil
	case "bubble":
		return Coral{coral(2), colour.Purple()}, nil
	case "fire":
		return Coral{coral(3), colour.Red()}, nil
	case "horn":
		return Coral{coral(4), colour.Yellow()}, nil
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
