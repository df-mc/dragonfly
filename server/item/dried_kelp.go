package item

// Dried kelp is a food item that can be quickly eaten by the player.
type DriedKelp struct{}

// EncodeItem ...
func (DriedKelp) EncodeItem() (name string, meta int16) {
	return "minecraft:dried_kelp", 0
}
