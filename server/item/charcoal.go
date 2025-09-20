package item

import "time"

// Charcoal is an item obtained by smelting logs or wood.
type Charcoal struct{}

func (Charcoal) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 80)
}

func (Charcoal) EncodeItem() (name string, meta int16) {
	return "minecraft:charcoal", 0
}
