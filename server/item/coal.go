package item

import "time"

// Coal is an item used as fuel & crafting torches.
type Coal struct{}

// FuelInfo ...
func (Coal) FuelInfo() FuelInfo {
	return FuelInfo{Duration: time.Second * 80}
}

// EncodeItem ...
func (Coal) EncodeItem() (name string, meta int16) {
	return "minecraft:coal", 0
}
