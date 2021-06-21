package item

// DragonBreath is a brewing item that is used solely to make lingering potions.
type DragonBreath struct{}

// EncodeItem ...
func (DragonBreath) EncodeItem() (name string, meta int16) {
	return "minecraft:dragon_breath", 0
}
