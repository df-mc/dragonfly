package item

// PrismarineCrystals are items obtained by defeating guardians or elder guardians. They are used for crafting sea
// lanterns.
type PrismarineCrystals struct{}

// EncodeItem ...
func (p PrismarineCrystals) EncodeItem() (name string, meta int16) {
	return "minecraft:prismarine_crystals", 0
}
