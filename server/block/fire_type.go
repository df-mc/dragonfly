package block

import "fmt"

// FireType represents a type of fire. Used by flaming blocks such as torches, lanterns, fire, and campfires.
type FireType struct {
	fire
}

type fire uint8

// NormalFire is the default variant of fires
func NormalFire() FireType {
	return FireType{fire(0)}
}

// SoulFire is a turquoise variant of normal fire
func SoulFire() FireType {
	return FireType{fire(1)}
}

// Uint8 returns the fire as a uint8.
func (f fire) Uint8() uint8 {
	return uint8(f)
}

// LightLevel returns the light level of the fire.
func (f fire) LightLevel() uint8 {
	switch f {
	case 0:
		return 15
	case 1:
		return 10
	}
	panic("unknown fire type")
}

// Name ...
func (f fire) Name() string {
	switch f {
	case 0:
		return "Fire"
	case 1:
		return "Soul Fire"
	}
	panic("unknown fire type")
}

// FromString ...
func (f fire) FromString(s string) (interface{}, error) {
	switch s {
	case "normal":
		return NormalFire(), nil
	case "soul":
		return SoulFire(), nil
	}
	return nil, fmt.Errorf("unexpected fire type '%v', expecting one of 'normal' or 'soul'", s)
}

// String ...
func (f fire) String() string {
	switch f {
	case 0:
		return "normal"
	case 1:
		return "soul"
	}
	panic("unknown fire type")
}

// FireTypes ...
func FireTypes() []FireType {
	return []FireType{NormalFire(), SoulFire()}
}
