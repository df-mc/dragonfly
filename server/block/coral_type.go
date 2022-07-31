package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// CoralType represents a type of coral of a block. CoralType, coral fans, and coral blocks carry one of these types.
type CoralType struct {
	coral
}

// TubeCoral returns the tube coral variant
func TubeCoral() CoralType {
	return CoralType{0}
}

// BrainCoral returns the brain coral variant
func BrainCoral() CoralType {
	return CoralType{1}
}

// BubbleCoral returns the bubble coral variant
func BubbleCoral() CoralType {
	return CoralType{2}
}

// FireCoral returns the fire coral variant
func FireCoral() CoralType {
	return CoralType{3}
}

// HornCoral returns the horn coral variant
func HornCoral() CoralType {
	return CoralType{4}
}

// CoralTypes returns all coral types.
func CoralTypes() []CoralType {
	return []CoralType{TubeCoral(), BrainCoral(), BubbleCoral(), FireCoral(), HornCoral()}
}

type coral uint8

// Uint8 returns the coral as a uint8.
func (c coral) Uint8() uint8 {
	return uint8(c)
}

// Colour returns the colour of the CoralType.
func (c coral) Colour() item.Colour {
	switch c {
	case 0:
		return item.ColourBlue()
	case 1:
		return item.ColourPink()
	case 2:
		return item.ColourPurple()
	case 3:
		return item.ColourRed()
	case 4:
		return item.ColourYellow()
	}
	panic("unknown coral type")
}

// Name ...
func (c coral) Name() string {
	switch c {
	case 0:
		return "Tube Coral"
	case 1:
		return "Brain Coral"
	case 2:
		return "Bubble Coral"
	case 3:
		return "Fire Coral"
	case 4:
		return "Horn Coral"
	}
	panic("unknown coral type")
}

// String ...
func (c coral) String() string {
	switch c {
	case 0:
		return "tube"
	case 1:
		return "brain"
	case 2:
		return "bubble"
	case 3:
		return "fire"
	case 4:
		return "horn"
	}
	panic("unknown coral type")
}
