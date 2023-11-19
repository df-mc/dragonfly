package block

// BellAttachment represents a type of attachment for a Bell.
type BellAttachment struct {
	bellAttachment
}

// StandingBellAttachment is a type of attachment for a standing Bell.
func StandingBellAttachment() BellAttachment {
	return BellAttachment{0}
}

// HangingBellAttachment is a type of attachment for a hanging Bell.
func HangingBellAttachment() BellAttachment {
	return BellAttachment{1}
}

// WallBellAttachment is a type of attachment for a wall Bell.
func WallBellAttachment() BellAttachment {
	return BellAttachment{2}
}

// WallsBellAttachment is a type of attachment for a two-wall Bell.
func WallsBellAttachment() BellAttachment {
	return BellAttachment{3}
}

// BellAttachments returns all possible BellAttachments.
func BellAttachments() []BellAttachment {
	return []BellAttachment{StandingBellAttachment(), HangingBellAttachment(), WallBellAttachment(), WallsBellAttachment()}
}

type bellAttachment uint8

// Uint8 returns the BellAttachment as a uint8.
func (g bellAttachment) Uint8() uint8 {
	return uint8(g)
}

// String returns the BellAttachment as a string.
func (g bellAttachment) String() string {
	switch g {
	case 0:
		return "standing"
	case 1:
		return "hanging"
	case 2:
		return "side"
	case 3:
		return "multiple"
	}
	panic("should never happen")
}
