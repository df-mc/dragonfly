package block

// MushroomType represents a type of mushroom. This is either brown or red.
type MushroomType struct {
	mushroom
}

// BrownMushroom returns the brown mushroom variant of mushrooms.
func BrownMushroom() MushroomType {
	return MushroomType{mushroom(0)}
}

// RedMushroom returns the red mushroom variant of mushrooms.
func RedMushroom() MushroomType {
	return MushroomType{mushroom(1)}
}

// MushroomTypes returns a list of all mushroom types.
func MushroomTypes() []MushroomType {
	return []MushroomType{BrownMushroom(), RedMushroom()}
}

type mushroom uint8

// Uint8 returns the mushroom as an uint8.
func (m mushroom) Uint8() uint8 {
	return uint8(m)
}

// String returns the mushroom as a string.
func (m mushroom) String() string {
	switch m {
	case 0:
		return "brown"
	case 1:
		return "red"
	}
	panic("unknown mushroom type")
}
