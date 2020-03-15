package material

// Wood represents a type of wood of a block. Some blocks, such as log blocks, bark blocks, wooden planks and
// others carry one of these types.
type Wood interface {
	// Name returns a human-readable name of the wood type, such as 'Oak'.
	Name() string
	// String returns a string that represents the wood type in Minecraft (over network, for example), such
	// as 'dark_oak'.
	String() string
	__()
}

// OakWood returns oak wood material.
func OakWood() Wood {
	return wood(0)
}

// SpruceWood returns spruce wood material.
func SpruceWood() Wood {
	return wood(1)
}

// BirchWood returns birch wood material.
func BirchWood() Wood {
	return wood(2)
}

// JungleWood returns jungle wood material.
func JungleWood() Wood {
	return wood(3)
}

// AcaciaWood returns acacia wood material.
func AcaciaWood() Wood {
	return wood(4)
}

// DarkOakWood returns dark oak wood material.
func DarkOakWood() Wood {
	return wood(5)
}

type wood uint8

func (w wood) __() {}

func (w wood) Name() string {
	switch w {
	case 0:
		return "Oak"
	case 1:
		return "Spruce"
	case 2:
		return "Birch"
	case 3:
		return "Jungle"
	case 4:
		return "Acacia"
	case 5:
		return "Dark Oak"
	}
	panic("unknown wood type")
}

func (w wood) String() string {
	switch w {
	case 0:
		return "oak"
	case 1:
		return "spruce"
	case 2:
		return "birch"
	case 3:
		return "jungle"
	case 4:
		return "acacia"
	case 5:
		return "dark_oak"
	}
	panic("unknown wood type")
}
