package item

import "time"

// Stick is one of the most abundant resources used for crafting many tools and items.
type Stick struct{}

// FuelInfo ...
func (Stick) FuelInfo() FuelInfo {
	return FuelInfo{Duration: time.Second * 5}
}

// EncodeItem ...
func (s Stick) EncodeItem() (name string, meta int16) {
	return "minecraft:stick", 0
}
