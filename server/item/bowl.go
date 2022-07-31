package item

import "time"

// Bowl is a container that can hold certain foods.
type Bowl struct{}

// FuelInfo ...
func (Bowl) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 10)
}

// EncodeItem ...
func (Bowl) EncodeItem() (name string, meta int16) {
	return "minecraft:bowl", 0
}
