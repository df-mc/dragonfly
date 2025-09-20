package item

// FireworkStar is an item used to determine the color, effect, and shape of firework rockets.
type FireworkStar struct {
	FireworkExplosion
}

func (f FireworkStar) EncodeNBT() map[string]any {
	return map[string]any{
		"FireworksItem": f.FireworkExplosion.EncodeNBT(),
		"customColor":   int32FromRGBA(f.Colour.RGBA()),
	}
}

func (f FireworkStar) DecodeNBT(data map[string]any) any {
	if i, ok := data["FireworksItem"].(map[string]any); ok {
		f.FireworkExplosion = f.FireworkExplosion.DecodeNBT(i).(FireworkExplosion)
	}
	return f
}

func (f FireworkStar) EncodeItem() (name string, meta int16) {
	return "minecraft:firework_star", invertColour(f.FireworkExplosion.Colour)
}
