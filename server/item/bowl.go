package item

import "time"

// Bowl is a container that can hold certain foods.
type Bowl struct{}

func (Bowl) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 10)
}

func (Bowl) EncodeItem() (name string, meta int16) {
	return "minecraft:bowl", 0
}
