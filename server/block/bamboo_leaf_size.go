package block

// BambooLeafSize represents the size of bamboo leaves.
type BambooLeafSize struct {
	bamboo
}

type bamboo uint8

// BambooSizeNoLeaves ...
func BambooSizeNoLeaves() BambooLeafSize {
	return BambooLeafSize{0}
}

// BambooSizeSmallLeaves ...
func BambooSizeSmallLeaves() BambooLeafSize {
	return BambooLeafSize{1}
}

// BambooSizeLargeLeaves ...
func BambooSizeLargeLeaves() BambooLeafSize {
	return BambooLeafSize{2}
}

// Uint8 ...
func (b bamboo) Uint8() uint8 {
	return uint8(b)
}

// String ...
func (b bamboo) String() string {
	switch b {
	case 0:
		return "no_leaves"
	case 1:
		return "small_leaves"
	case 2:
		return "large_leaves"
	}
	panic("unknown bamboo leaf size")
}

// Name ...
func (b bamboo) Name() string {
	switch b {
	case 0:
		return "No Leaves"
	case 1:
		return "Small Leaves"
	case 2:
		return "Large Leaves"
	}
	panic("unknown bamboo leaf size")
}

// BambooLeafSizes returns all possible bamboo leaf sizes.
func BambooLeafSizes() []BambooLeafSize {
	return []BambooLeafSize{BambooSizeNoLeaves(), BambooSizeSmallLeaves(), BambooSizeLargeLeaves()}
}
