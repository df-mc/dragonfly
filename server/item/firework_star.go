package item

import colourconv "github.com/df-mc/dragonfly/server/internal/colour"

// FireworkStar is an item used to determine the color, effect, and shape of firework rockets.
type FireworkStar struct {
	FireworkExplosion
}

// EncodeNBT ...
func (f FireworkStar) EncodeNBT() map[string]any {
	return map[string]any{
		"FireworksItem": f.FireworkExplosion.EncodeNBT(),
		"customColor":   colourconv.Int32FromRGBAOpaqueBlack(f.Colour.RGBA()),
	}
}

// DecodeNBT ...
func (f FireworkStar) DecodeNBT(data map[string]any) any {
	if i, ok := data["FireworksItem"].(map[string]any); ok {
		f.FireworkExplosion = f.FireworkExplosion.DecodeNBT(i).(FireworkExplosion)
	}
	return f
}

// EncodeItem ...
func (f FireworkStar) EncodeItem() (name string, meta int16) {
	return "minecraft:firework_star", invertColour(f.Colour)
}
