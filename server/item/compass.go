package item

// Compass is an item used to find the spawn position of a world.
type Compass struct {
	// TrackingHandle is the handle of the lodestone tracked by the compass. A
	// value of 0 makes the compass point to the world spawn instead.
	TrackingHandle int32
}

// EncodeItem ...
func (c Compass) EncodeItem() (name string, meta int16) {
	if c.TrackingHandle != 0 {
		return "minecraft:lodestone_compass", 0
	}
	return "minecraft:compass", 0
}

// Glinted returns true if the compass is linked to a lodestone.
func (c Compass) Glinted() bool { return c.TrackingHandle != 0 }

// EncodeNBT encodes the position tracking handle understood by the Bedrock client.
func (c Compass) EncodeNBT() map[string]any {
	return map[string]any{"trackingHandle": c.TrackingHandle}
}

// DecodeNBT decodes a lodestone compass from NBT.
func (c Compass) DecodeNBT(data map[string]any) any {
	switch handle := data["trackingHandle"].(type) {
	case int32:
		c.TrackingHandle = handle
	case int:
		c.TrackingHandle = int32(handle)
	}
	return c
}
