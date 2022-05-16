package item

import "time"

// BlazeRod is an item exclusively obtained from blazes.
type BlazeRod struct{}

// FuelInfo ...
func (BlazeRod) FuelInfo() FuelInfo {
	return FuelInfo{Duration: time.Second * 120}
}

// EncodeItem ...
func (BlazeRod) EncodeItem() (name string, meta int16) {
	return "minecraft:blaze_rod", 0
}
