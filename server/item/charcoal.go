package item

import "time"

// Charcoal is an item obtained by smelting logs or wood.
type Charcoal struct{}

// FuelInfo ...
func (Charcoal) FuelInfo() FuelInfo {
	return FuelInfo{Duration: time.Second * 80}
}

// EncodeItem ...
func (Charcoal) EncodeItem() (name string, meta int16) {
	return "minecraft:charcoal", 0
}
