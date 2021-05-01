package item

// PrismarineCrystals are items obtained by defeating guardians or elder guardians. They are used to craft sea lanterns.
type PrismarineCrystals struct{}

// EncodeItem ...
func (p PrismarineCrystals) EncodeItem() (id int32, name string, meta int16) {
	return 422, "minecraft:prismarine_crystals", 0
}
