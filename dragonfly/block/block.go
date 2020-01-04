package block

// Block is a block that may be placed or found in a world. In addition, the block may also be added to an
// inventory: It is also an item.
type Block interface {
	// Name returns the readable name of the block. An example for oak log would be 'Oak Log'.
	Name() string

	// Minecraft converts the block to its Minecraft representation: It returns the name of the minecraft
	// block, for example 'minecraft:stone' and the block properties (also referred to as states) that the
	// block holds.
	Minecraft() (name string, properties map[string]interface{})
}
