package item

// LapisLazuli is a mineral used for enchanting and decoration.
type LapisLazuli struct{}

// EncodeItem ...
func (LapisLazuli) EncodeItem() (name string, meta int16) {
	return "minecraft:lapis_lazuli", 0
}
