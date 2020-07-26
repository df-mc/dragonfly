package fire

import "fmt"

type Fire struct {
	fire
	Light uint8
}

type fire uint8

// Normal is the default variant of fires
func Normal() Fire {
	return Fire{fire(0), 15}
}

// Soul is a turquoise variant of normal fire
func Soul() Fire {
	return Fire{fire(1), 10}
}

// Uint8 returns the fire as a uint8.
func (f fire) Uint8() uint8 {
	return uint8(f)
}

// Name ...
func (f fire) Name() string {
	switch f {
	case 0:
		return "Normal"
	case 1:
		return "Soul"
	}
	panic("unknown wood type")
}

// FromString ...
func (f fire) FromString(s string) (interface{}, error) {
	switch s {
	case "normal":
		return Fire{fire(0), 15}, nil
	case "soul":
		return Fire{fire(1), 10}, nil
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
