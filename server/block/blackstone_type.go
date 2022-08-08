package block

// BlackstoneType represents a type of blackstone.
type BlackstoneType struct {
	blackstone
}

type blackstone uint8

// NormalBlackstone is the normal variant of blackstone.
func NormalBlackstone() BlackstoneType {
	return BlackstoneType{0}
}

// GildedBlackstone is the gilded variant of blackstone.
func GildedBlackstone() BlackstoneType {
	return BlackstoneType{1}
}

// PolishedBlackstone is the polished variant of blackstone.
func PolishedBlackstone() BlackstoneType {
	return BlackstoneType{2}
}

// ChiseledPolishedBlackstone is the chiseled polished variant of blackstone.
func ChiseledPolishedBlackstone() BlackstoneType {
	return BlackstoneType{3}
}

// Uint8 returns the blackstone type as a uint8.
func (s blackstone) Uint8() uint8 {
	return uint8(s)
}

// Name ...
func (s blackstone) Name() string {
	switch s {
	case 0:
		return "Blackstone"
	case 1:
		return "Gilded Blackstone"
	case 2:
		return "Polished Blackstone"
	case 3:
		return "Chiseled Polished Blackstone"
	}
	panic("unknown blackstone type")
}

// String ...
func (s blackstone) String() string {
	switch s {
	case 0:
		return "blackstone"
	case 1:
		return "gilded_blackstone"
	case 2:
		return "polished_blackstone"
	case 3:
		return "chiseled_polished_blackstone"
	}
	panic("unknown blackstone type")
}

// BlackstoneTypes ...
func BlackstoneTypes() []BlackstoneType {
	return []BlackstoneType{NormalBlackstone(), GildedBlackstone(), PolishedBlackstone(), ChiseledPolishedBlackstone()}
}
