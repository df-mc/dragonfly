package item

// LapisLazuli is a mineral used for enchanting and decoration.
type LapisLazuli struct{}

// EncodeItem ...
func (LapisLazuli) EncodeItem() (id int32, name string, meta int16) {
	return 351, "minecraft:dye", 4
}
