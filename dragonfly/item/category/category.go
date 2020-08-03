package category

type Category struct {
	category
}

func Construction() Category {
	return Category{"Construction"}
}

func Nature() Category {
	return Category{"Nature"}
}

func Equipment() Category {
	return Category{"Equipment"}
}

func Items() Category {
	return Category{"Items"}
}

type category string

func (c category) String() string {
	return string(c)
}
