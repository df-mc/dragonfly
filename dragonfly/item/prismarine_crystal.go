package item

// Prismarine Crystals is a Building Material to build Sealanterns, it can be collected by Killing Guardians or breaking Sealantern Blocks.
type PrismarineCrystals struct{}

// EncodeItem ...
func (PrismarineCrystals) EncodeItem() (id int32, meta int16) {
	return 422, 0
}
