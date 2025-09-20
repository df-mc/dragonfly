package item

import "time"

// BlazeRod is an item exclusively obtained from blazes.
type BlazeRod struct{}

func (BlazeRod) FuelInfo() FuelInfo {
	return newFuelInfo(time.Second * 120)
}

func (BlazeRod) EncodeItem() (name string, meta int16) {
	return "minecraft:blaze_rod", 0
}
