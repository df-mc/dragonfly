package block

// BellAttachment represents a type of attachment for a Bell.
type BellAttachment struct {
	bellAttachment
}

// StandingBellAttachment returns a floor-mounted bell attachment.
func StandingBellAttachment() BellAttachment {
	return BellAttachment{0}
}

// HangingBellAttachment returns a ceiling-mounted bell attachment.
func HangingBellAttachment() BellAttachment {
	return BellAttachment{1}
}

// SideBellAttachment returns a wall-mounted bell attachment with a single support.
func SideBellAttachment() BellAttachment {
	return BellAttachment{2}
}

// MultipleBellAttachment returns a wall-mounted bell attachment with supports on both sides.
func MultipleBellAttachment() BellAttachment {
	return BellAttachment{3}
}

// BellAttachments returns all possible Bell attachments.
func BellAttachments() []BellAttachment {
	return []BellAttachment{StandingBellAttachment(), HangingBellAttachment(), SideBellAttachment(), MultipleBellAttachment()}
}

type bellAttachment uint8

// Uint8 returns the BellAttachment as a uint8.
func (b bellAttachment) Uint8() uint8 {
	return uint8(b)
}

// String returns the BellAttachment as a string.
func (b bellAttachment) String() string {
	switch b {
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
