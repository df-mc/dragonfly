package category

// Category is used to categorize groups of creative items client-side.
type Category struct {
	group    string
	category uint8
}

// Construction ...
func Construction() Category {
	return Category{category: 1}
}

// Nature ...
func Nature() Category {
	return Category{category: 2}
}

// Equipment ...
func Equipment() Category {
	return Category{category: 3}
}

// Items ...
func Items() Category {
	return Category{category: 4}
}

// Uint8 ...
func (c Category) Uint8() uint8 {
	return c.category
}

// WithGroup ...
func (c Category) WithGroup(group string) Category {
	c.group = group
	return c
}

// Name ...
func (c Category) Name() string {
	switch c.category {
	case 1:
		return "construction"
	case 2:
		return "nature"
	case 3:
		return "equipment"
	case 4:
		return "items"
	}
	panic("should never happen")
}

// String ...
func (c Category) String() string {
	if len(c.group) > 0 {
		return "itemGroup.name." + c.group
	}
	return "itemGroup." + c.Name() + ".name"
}
