package item

import "github.com/sandertv/gophertunnel/minecraft/text"

// Emerald is a rare mineral obtained from emerald ore or from villagers.
type Emerald struct{}

// EncodeItem ...
func (Emerald) EncodeItem() (name string, meta int16) {
	return "minecraft:emerald", 0
}

// TrimMaterial ...
func (Emerald) TrimMaterial() string {
	return "emerald"
}

// MaterialColour ...
func (Emerald) MaterialColour() string {
	return text.Emerald
}

// PayableForBeacon ...
func (Emerald) PayableForBeacon() bool {
	return true
}
