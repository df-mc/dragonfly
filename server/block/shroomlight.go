package block

// Shroomlight are light-emitting blocks that generate in huge fungi.
type Shroomlight struct {
	solid
}

// LightEmissionLevel ...
func (Shroomlight) LightEmissionLevel() uint8 {
	return 15
}

// BreakInfo ...
func (s Shroomlight) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, hoeEffective, oneOf(s))
}

// EncodeItem ...
func (Shroomlight) EncodeItem() (name string, meta int16) {
	return "minecraft:shroomlight", 0
}

// EncodeBlock ...
func (Shroomlight) EncodeBlock() (string, map[string]any) {
	return "minecraft:shroomlight", nil
}
