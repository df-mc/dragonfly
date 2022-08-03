package item

// FireworkStar is an item used to determine the color, effect, and shape of firework rockets.
type FireworkStar struct {
	FireworkExplosion
}

// EncodeItemNBT ...
func (f FireworkStar) EncodeItemNBT() map[string]any {
	return map[string]any{
		"FireworksItem": f.FireworkExplosion.EncodeNBT(),
		"customColor":   int32FromRGBA(f.Colour.RGBA()),
	}
}

// DecodeItemNBT ...
func (f FireworkStar) DecodeItemNBT(data map[string]any) any {
	f.FireworkExplosion = f.FireworkExplosion.DecodeNBT(data["FireworksItem"].(map[string]any)).(FireworkExplosion)
	return f
}

// EncodeItem ...
func (f FireworkStar) EncodeItem() (name string, meta int16) {
	return "minecraft:firework_star", invertColour(f.Colour)
}
