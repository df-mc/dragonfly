package block

import "fmt"

// FlowerType represents a type of flower.
type FlowerType struct {
	flower
}

type flower uint8

// Dandelion is a dandelion flower.
func Dandelion() FlowerType {
	return FlowerType{flower(0)}
}

// Poppy is a poppy flower.
func Poppy() FlowerType {
	return FlowerType{flower(1)}
}

// BlueOrchid is a blue orchid flower.
func BlueOrchid() FlowerType {
	return FlowerType{flower(2)}
}

// Allium is a allium flower.
func Allium() FlowerType {
	return FlowerType{flower(3)}
}

// AzureBluet is an azure bluet flower.
func AzureBluet() FlowerType {
	return FlowerType{flower(4)}
}

// RedTulip is a red tulip flower.
func RedTulip() FlowerType {
	return FlowerType{flower(5)}
}

// OrangeTulip is an orange tulip flower.
func OrangeTulip() FlowerType {
	return FlowerType{flower(6)}
}

// WhiteTulip is a white tulip flower.
func WhiteTulip() FlowerType {
	return FlowerType{flower(7)}
}

// PinkTulip is a pink tulip flower.
func PinkTulip() FlowerType {
	return FlowerType{flower(8)}
}

// OxeyeDaisy is an oxeye daisy flower.
func OxeyeDaisy() FlowerType {
	return FlowerType{flower(9)}
}

// Cornflower is a cornflower flower.
func Cornflower() FlowerType {
	return FlowerType{flower(10)}
}

// LilyOfTheValley is a lily of the valley flower.
func LilyOfTheValley() FlowerType {
	return FlowerType{flower(11)}
}

// WitherRose is a wither rose flower.
func WitherRose() FlowerType {
	return FlowerType{flower(12)}
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

// FromString ...
func (f flower) FromString(s string) (interface{}, error) {
	switch s {
	case "dandelion":
		return FlowerType{flower(0)}, nil
	case "poppy":
		return FlowerType{flower(1)}, nil
	case "orchid":
		return FlowerType{flower(2)}, nil
	case "allium":
		return FlowerType{flower(3)}, nil
	case "houstonia":
		return FlowerType{flower(4)}, nil
	case "tulip_red":
		return FlowerType{flower(5)}, nil
	case "tulip_orange":
		return FlowerType{flower(6)}, nil
	case "tulip_white":
		return FlowerType{flower(7)}, nil
	case "tulip_pink":
		return FlowerType{flower(8)}, nil
	case "oxeye":
		return FlowerType{flower(9)}, nil
	case "cornflower":
		return FlowerType{flower(10)}, nil
	case "lily_of_the_valley":
		return FlowerType{flower(11)}, nil
	case "wither_rose":
		return FlowerType{flower(12)}, nil
	}
	return nil, fmt.Errorf("unexpected flower type '%v', expecting one of 'dandelion', 'poppy', 'orchid', 'allium', 'houstonia', 'tulip_red', 'tulip_orange', 'tulip_white', 'tulip_pink', 'oxeye', 'cornflower', 'lily_of_the_valley', or 'wither_rose'", s)
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
