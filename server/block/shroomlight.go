package block

// Shroomlight are light-emitting blocks that generate in huge fungi.
type Shroomlight struct {
	solid
}

func (Shroomlight) LightEmissionLevel() uint8 {
	return 15
}

func (s Shroomlight) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, hoeEffective, oneOf(s))
}

func (Shroomlight) CompostChance() float64 {
	return 0.65
}

func (Shroomlight) EncodeItem() (name string, meta int16) {
	return "minecraft:shroomlight", 0
}

func (Shroomlight) EncodeBlock() (string, map[string]any) {
	return "minecraft:shroomlight", nil
}
