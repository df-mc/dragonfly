package block

import "fmt"

type PurpurType struct {
	purpur
}

type purpur uint8

// NormalPurpur is the normal variant of purpur.
func NormalPurpur() PurpurType {
	return PurpurType{purpur(0)}
}

// ChiseledPurpur is the chiseled variant of purpur. It is only accessible only through commands.
func ChiseledPurpur() PurpurType {
	return PurpurType{purpur(1)}
}

// PillarPurpur is the pillar variant of purpur.
func PillarPurpur() PurpurType {
	return PurpurType{purpur(2)}
}

// SmoothPurpur is the smooth variant of purpur. It is only accessible only through commands.
func SmoothPurpur() PurpurType {
	return PurpurType{purpur(3)}
}

// Uint8 returns the purpur as a uint8.
func (p purpur) Uint8() uint8 {
	return uint8(p)
}

// Name ...
func (p purpur) Name() string {
	switch p {
	case 0:
		return "Purpur"
	case 1:
		return "Chiseled Purpur"
	case 2:
		return "Purpur Pillar"
	case 3:
		return "Smooth Purpur"
	}
	panic("unknown purpur type")
}

// FromString ...
func (p purpur) FromString(str string) (interface{}, error) {
	switch str {
	case "default", "normal":
		return NormalPurpur(), nil
	case "chiseled":
		return ChiseledPurpur(), nil
	case "lines":
		return PillarPurpur(), nil
	case "smooth":
		return SmoothPurpur(), nil
	}
	return nil, fmt.Errorf("unexpected purpur type '%v'", p)
}

// String ...
func (p purpur) String() string {
	switch p {
	case 0:
		return "default"
	case 1:
		return "chiseled"
	case 2:
		return "lines"
	case 3:
		return "smooth"
	}
	panic("unknown purpur type")
}

// PurpurTypes returns a list of all purpur types.
func PurpurTypes() []PurpurType {
	return []PurpurType{NormalPurpur(), ChiseledPurpur(), PillarPurpur(), SmoothPurpur()}
}
