package item

// Slimeball is a crafting ingredient commonly dropped by slimes, and can be sneezed out by pandas.
type Slimeball struct{}

func (Slimeball) EncodeItem() (name string, meta int16) {
	return "minecraft:slime_ball", 0
}
