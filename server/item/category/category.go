package category

// Category represents the category a custom item will be displayed in, in the creative inventory.
type Category struct {
	group    string
	category uint8
}

// Construction is the first tab in the creative inventory and usually contains blocks that are more for decoration and
// building than actual functionality.
func Construction() Category {
	return Category{category: 1}
}

// Nature is the fourth tab in the creative inventory and usually contains blocks and items that can be found naturally
// in vanilla-generated world.
func Nature() Category {
	return Category{category: 2}
}

// Equipment is the second tab in the creative inventory and usually contains armour, weapons and tools.
func Equipment() Category {
	return Category{category: 3}
}

// Items is the third tab in the creative inventory and usually contains blocks and items that do not come under any
// other category, such as minerals, mob drops, containers and redstone etc.
func Items() Category {
	return Category{category: 4}
}

// Uint8 ...
func (c Category) Uint8() uint8 {
	return c.category
}

// WithGroup returns the category with the provided subgroup. This can be used to put an item inside a group such as
// swords, food or different types of blocks.
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
