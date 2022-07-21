package block

// FlowerType represents a type of flower.
type FlowerType struct {
	flower
}

type flower uint8

// Dandelion is a dandelion flower.
func Dandelion() FlowerType {
	return FlowerType{0}
}

// Poppy is a poppy flower.
func Poppy() FlowerType {
	return FlowerType{1}
}

// BlueOrchid is a blue orchid flower.
func BlueOrchid() FlowerType {
	return FlowerType{2}
}

// Allium is an allium flower.
func Allium() FlowerType {
	return FlowerType{3}
}

// AzureBluet is an azure bluet flower.
func AzureBluet() FlowerType {
	return FlowerType{4}
}

// RedTulip is a red tulip flower.
func RedTulip() FlowerType {
	return FlowerType{5}
}

// OrangeTulip is an orange tulip flower.
func OrangeTulip() FlowerType {
	return FlowerType{6}
}

// WhiteTulip is a white tulip flower.
func WhiteTulip() FlowerType {
	return FlowerType{7}
}

// PinkTulip is a pink tulip flower.
func PinkTulip() FlowerType {
	return FlowerType{8}
}

// OxeyeDaisy is an oxeye daisy flower.
func OxeyeDaisy() FlowerType {
	return FlowerType{9}
}

// Cornflower is a cornflower flower.
func Cornflower() FlowerType {
	return FlowerType{10}
}

// LilyOfTheValley is a lily of the valley flower.
func LilyOfTheValley() FlowerType {
	return FlowerType{11}
}

// WitherRose is a wither rose flower.
func WitherRose() FlowerType {
	return FlowerType{12}
}

// Uint8 returns the flower as a uint8.
func (f flower) Uint8() uint8 {
	return uint8(f)
}

// Name ...
func (f flower) Name() string {
	switch f {
	case 0:
		return "Dandelion"
	case 1:
		return "Poppy"
	case 2:
		return "Blue Orchid"
	case 3:
		return "Allium"
	case 4:
		return "Azure Bluet"
	case 5:
		return "Red Tulip"
	case 6:
		return "Orange Tulip"
	case 7:
		return "White Tulip"
	case 8:
		return "Pink Tulip"
	case 9:
		return "Oxeye Daisy"
	case 10:
		return "Cornflower"
	case 11:
		return "Lily of the Valley"
	case 12:
		return "Wither Rose"
	}
	panic("unknown flower type")
}

// String ...
func (f flower) String() string {
	switch f {
	case 0:
		return "dandelion"
	case 1:
		return "poppy"
	case 2:
		return "orchid"
	case 3:
		return "allium"
	case 4:
		return "houstonia"
	case 5:
		return "tulip_red"
	case 6:
		return "tulip_orange"
	case 7:
		return "tulip_white"
	case 8:
		return "tulip_pink"
	case 9:
		return "oxeye"
	case 10:
		return "cornflower"
	case 11:
		return "lily_of_the_valley"
	case 12:
		return "wither_rose"
	}
	panic("unknown flower type")
}

// FlowerTypes ...
func FlowerTypes() []FlowerType {
	return []FlowerType{Dandelion(), Poppy(), BlueOrchid(), Allium(), AzureBluet(), RedTulip(), OrangeTulip(), WhiteTulip(), PinkTulip(), OxeyeDaisy(), Cornflower(), LilyOfTheValley(), WitherRose()}
}
