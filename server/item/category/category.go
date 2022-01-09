package category

// Category is used to categorize groups of creative items client-side.
type Category struct {
	category
}

// Construction ...
func Construction() Category {
	return Category{2}
}

// Nature ...
func Nature() Category {
	return Category{3}
}

// Equipment ...
func Equipment() Category {
	return Category{4}
}

// Items ...
func Items() Category {
	return Category{5}
}

type category uint8

// Uint8 ...
func (c category) Uint8() uint8 {
	return uint8(c)
}

// String ...
func (c category) String() string {
	switch c {
	case 2:
		return "construction"
	case 3:
		return "nature"
	case 4:
		return "equipment"
	case 5:
		return "items"
	}
	panic("should never happen")
}
