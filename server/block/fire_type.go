package block

// FireType represents a type of fire. Used by flaming blocks such as torches, lanterns, fire, and campfires.
type FireType struct {
	fire
}

type fire uint8

// NormalFire is the default variant of fires
func NormalFire() FireType {
	return FireType{0}
}

// SoulFire is a turquoise variant of normal fire
func SoulFire() FireType {
	return FireType{1}
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

// Damage returns the amount of damage taken by entities inside the fire.
func (f fire) Damage() float64 {
	switch f {
	case 0:
		return 1
	case 1:
		return 2
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
