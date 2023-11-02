package item

// SherdType represents the sherdType of a block.
type SherdType struct {
	sherdType
}

// SherdTypeAngler returns the angler sherd type.
func SherdTypeAngler() SherdType {
	return SherdType{0}
}

// SherdTypeArcher returns the archer sherd type.
func SherdTypeArcher() SherdType {
	return SherdType{1}
}

// SherdTypeArmsUp returns the arms up sherd type.
func SherdTypeArmsUp() SherdType {
	return SherdType{2}
}

// SherdTypeBlade returns the blade sherd type.
func SherdTypeBlade() SherdType {
	return SherdType{3}
}

// SherdTypeBrewer returns the brewer sherd type.
func SherdTypeBrewer() SherdType {
	return SherdType{4}
}

// SherdTypeBurn returns the burn sherd type.
func SherdTypeBurn() SherdType {
	return SherdType{5}
}

// SherdTypeDanger returns the danger sherd type.
func SherdTypeDanger() SherdType {
	return SherdType{6}
}

// SherdTypeExplorer returns the explorer sherd type.
func SherdTypeExplorer() SherdType {
	return SherdType{7}
}

// SherdTypeFriend returns the friend sherd type.
func SherdTypeFriend() SherdType {
	return SherdType{8}
}

// SherdTypeHeart returns the heart sherd type.
func SherdTypeHeart() SherdType {
	return SherdType{9}
}

// SherdTypeHeartbreak returns the heartbreak sherd type.
func SherdTypeHeartbreak() SherdType {
	return SherdType{10}
}

// SherdTypeHowl returns the howl sherd type.
func SherdTypeHowl() SherdType {
	return SherdType{11}
}

// SherdTypeMiner returns the miner sherd type.
func SherdTypeMiner() SherdType {
	return SherdType{12}
}

// SherdTypeMourner returns the mourner sherd type.
func SherdTypeMourner() SherdType {
	return SherdType{13}
}

// SherdTypePlenty returns the plenty sherd type.
func SherdTypePlenty() SherdType {
	return SherdType{14}
}

// SherdTypePrize returns the prize sherd type.
func SherdTypePrize() SherdType {
	return SherdType{15}
}

// SherdTypeSheaf returns the sheaf sherd type.
func SherdTypeSheaf() SherdType {
	return SherdType{16}
}

// SherdTypeShelter returns the shelter sherd type.
func SherdTypeShelter() SherdType {
	return SherdType{17}
}

// SherdTypeSkull returns the skull sherd type.
func SherdTypeSkull() SherdType {
	return SherdType{18}
}

// SherdTypeSnort returns the snort sherd type.
func SherdTypeSnort() SherdType {
	return SherdType{19}
}

// SherdTypes returns a list of all existing sherd types.
func SherdTypes() []SherdType {
	return []SherdType{
		SherdTypeAngler(), SherdTypeArcher(), SherdTypeArmsUp(), SherdTypeBlade(), SherdTypeBrewer(), SherdTypeBurn(),
		SherdTypeDanger(), SherdTypeExplorer(), SherdTypeFriend(), SherdTypeHeart(), SherdTypeHeartbreak(), SherdTypeHowl(),
		SherdTypeMiner(), SherdTypeMourner(), SherdTypePlenty(), SherdTypePrize(), SherdTypeSheaf(), SherdTypeShelter(),
		SherdTypeSkull(), SherdTypeSnort(),
	}
}

// sherdType is the underlying value of a SherdType struct.
type sherdType uint8

// String ...
func (c sherdType) String() string {
	switch c {
	case 0:
		return "angler"
	case 1:
		return "archer"
	case 2:
		return "arms_up"
	case 3:
		return "blade"
	case 4:
		return "brewer"
	case 5:
		return "burn"
	case 6:
		return "danger"
	case 7:
		return "explorer"
	case 8:
		return "friend"
	case 9:
		return "heart"
	case 10:
		return "heartbreak"
	case 11:
		return "howl"
	case 12:
		return "miner"
	case 13:
		return "mourner"
	case 14:
		return "plenty"
	case 15:
		return "prize"
	case 16:
		return "sheaf"
	case 17:
		return "shelter"
	case 18:
		return "skull"
	case 19:
		return "snort"
	}
	panic("unknown sherd type")
}

// Uint8 ...
func (c sherdType) Uint8() uint8 {
	return uint8(c)
}
