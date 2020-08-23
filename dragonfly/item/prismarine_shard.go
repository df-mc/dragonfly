package item

// Prismarine Shard is a Building Material to build Prismarine Blocks, it can be collected by Killing Guardians.
type PrismarineShard struct{}

// EncodeItem ...
func (PrismarineShard) EncodeItem() (id int32, meta int16) {
	return 409, 0
}
