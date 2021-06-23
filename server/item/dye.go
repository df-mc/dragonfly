package item

import "github.com/df-mc/dragonfly/server/world"

// Dyes are made up of 16 different type of colours which allows you to dye blocks like concrete and sheep

// BlackDye ...
type BlackDye struct{}

// BlueDye ...
type BlueDye struct{}

// BrownDye ...
type BrownDye struct{}

// CyanDye ...
type CyanDye struct{}

// GrayDye ...
type GrayDye struct{}

// GreenDye ...
type GreenDye struct{}

// LightBlueDye ...
type LightBlueDye struct{}

// LightGrayDye ...
type LightGrayDye struct{}

// LimeDye ...
type LimeDye struct{}

// MagentaDye ...
type MagentaDye struct{}

// OrangeDye ...
type OrangeDye struct{}

// PinkDye ...
type PinkDye struct{}

// PurpleDye ...
type PurpleDye struct{}

// RedDye ...
type RedDye struct{}

// WhiteDye ...
type WhiteDye struct{}

// YellowDye ...
type YellowDye struct{}

// EncodeItem ...
func (BlackDye) EncodeItem() (name string, meta int16) {
	return "minecraft:black_dye", 0
}

// EncodeItem ...
func (BlueDye) EncodeItem() (name string, meta int16) {
	return "minecraft:blue_dye", 0
}

// EncodeItem ...
func (BrownDye) EncodeItem() (name string, meta int16) {
	return "minecraft:brown_dye", 0
}

// EncodeItem ...
func (CyanDye) EncodeItem() (name string, meta int16) {
	return "minecraft:cyan_dye", 0
}

// EncodeItem ...
func (GrayDye) EncodeItem() (name string, meta int16) {
	return "minecraft:gray_dye", 0
}

// EncodeItem ...
func (GreenDye) EncodeItem() (name string, meta int16) {
	return "minecraft:green_dye", 0
}

// EncodeItem ...
func (LightBlueDye) EncodeItem() (name string, meta int16) {
	return "minecraft:light_blue_dye", 0
}

// EncodeItem ...
func (LightGrayDye) EncodeItem() (name string, meta int16) {
	return "minecraft:light_gray_dye", 0
}

// EncodeItem ...
func (LimeDye) EncodeItem() (name string, meta int16) {
	return "minecraft:lime_dye", 0
}

// EncodeItem ...
func (MagentaDye) EncodeItem() (name string, meta int16) {
	return "minecraft:magenta_dye", 0
}

// EncodeItem ...
func (OrangeDye) EncodeItem() (name string, meta int16) {
	return "minecraft:orange_dye", 0
}

// EncodeItem ...
func (PinkDye) EncodeItem() (name string, meta int16) {
	return "minecraft:pink_dye", 0
}

// EncodeItem ...
func (PurpleDye) EncodeItem() (name string, meta int16) {
	return "minecraft:purple_dye", 0
}

// EncodeItem ...
func (RedDye) EncodeItem() (name string, meta int16) {
	return "minecraft:red_dye", 0
}

// EncodeItem ...
func (WhiteDye) EncodeItem() (name string, meta int16) {
	return "minecraft:white_dye", 0
}

// EncodeItem ...
func (YellowDye) EncodeItem() (name string, meta int16) {
	return "minecraft:yellow_dye", 0
}

// AllDyes returns all 16 dye items
func AllDyes() []world.Item {
	return []world.Item{
		BlackDye{}, BlueDye{}, BrownDye{}, CyanDye{}, GrayDye{}, GreenDye{}, LightBlueDye{}, LightGrayDye{},
		LimeDye{}, MagentaDye{}, OrangeDye{}, PinkDye{}, PurpleDye{}, RedDye{}, WhiteDye{}, YellowDye{},
	}
}
