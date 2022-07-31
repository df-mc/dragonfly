package block

// OreType represents a variant of ore blocks.
type OreType struct {
	ore
}

// StoneOre returns the stone ore variant.
func StoneOre() OreType {
	return OreType{0}
}

// DeepslateOre returns the deepslate ore variant.
func DeepslateOre() OreType {
	return OreType{1}
}

// OreTypes returns a list of all ore types
func OreTypes() []OreType {
	return []OreType{StoneOre(), DeepslateOre()}
}

type ore uint8

// Uint8 returns the ore as a uint8.
func (o ore) Uint8() uint8 {
	return uint8(o)
}

// Name ...
func (o ore) Name() string {
	switch o {
	case 0:
		return "Stone"
	case 1:
		return "Deepslate"
	}
	panic("unknown ore type")
}

// String ...
func (o ore) String() string {
	switch o {
	case 0:
		return "stone"
	case 1:
		return "deepslate"
	}
	panic("unknown ore type")
}

// Prefix ...
func (o ore) Prefix() string {
	switch o {
	case 0:
		return ""
	case 1:
		return "deepslate_"
	}
	panic("unknown ore type")
}

// Hardness ...
func (o ore) Hardness() float64 {
	switch o {
	case 0:
		return 3
	case 1:
		return 4.5
	}
	panic("unknown ore type")
}
