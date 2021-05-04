package item

// Emerald is a rare mineral obtained from emerald ore or from villagers.
type Emerald struct{}

// EncodeItem ...
func (Emerald) EncodeItem() (name string, meta int16) {
	return "minecraft:emerald", 0
}

// PayableForBeacon ...
func (Emerald) PayableForBeacon() bool {
	return true
}
