package block

import "fmt"

// DoublePlantType represents a type of double plant.
type DoublePlantType struct {
	doublePlant
}

type doublePlant uint8

// Sunflower is a sunflower plant.
func Sunflower() DoublePlantType {
	return DoublePlantType{doublePlant(0)}
}

// Lilac is a lilac plant.
func Lilac() DoublePlantType {
	return DoublePlantType{doublePlant(1)}
}

// RoseBush is a rose bush plant.
func RoseBush() DoublePlantType {
	return DoublePlantType{doublePlant(4)}
}

// Peony is a peony plant.
func Peony() DoublePlantType {
	return DoublePlantType{doublePlant(5)}
}

// Uint8 returns the double plant as a uint8.
func (d doublePlant) Uint8() uint8 {
	return uint8(d)
}

// Name ...
func (d doublePlant) Name() string {
	switch d {
	case 0:
		return "Sunflower"
	case 1:
		return "Lilac"
	case 4:
		return "Rose Bush"
	case 5:
		return "Peony"
	}
	panic("unknown double plant type")
}

// FromString ...
func (d doublePlant) FromString(s string) (interface{}, error) {
	switch s {
	case "sunflower":
		return DoublePlantType{doublePlant(0)}, nil
	case "syringa":
		return DoublePlantType{doublePlant(1)}, nil
	case "rose":
		return DoublePlantType{doublePlant(4)}, nil
	case "paeonia":
		return DoublePlantType{doublePlant(5)}, nil
	}
	return nil, fmt.Errorf("unexpected double plant type '%v', expecting one of 'sunflower', 'syringa', 'rose', or 'paeonia'", s)
}

// String ...
func (d doublePlant) String() string {
	switch d {
	case 0:
		return "sunflower"
	case 1:
		return "syringa"
	case 4:
		return "rose"
	case 5:
		return "paeonia"
	}
	panic("unknown double plant type")
}

// DoublePlantTypes ...
func DoublePlantTypes() []DoublePlantType {
	return []DoublePlantType{Sunflower(), Lilac(), RoseBush(), Peony()}
}
