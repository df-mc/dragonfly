package block

// Bedrock is a block that is indestructible in survival.
type Bedrock struct {
	solid
	transparent
	bassDrum

	// InfiniteBurning specifies if the bedrock block is set aflame and will burn forever. This is the case
	// for bedrock found under end crystals on top of the end pillars.
	InfiniteBurning bool
}

// EncodeItem ...
func (Bedrock) EncodeItem() (name string, meta int16) {
	return "minecraft:bedrock", 0
}

// EncodeBlock ...
func (b Bedrock) EncodeBlock() (name string, properties map[string]any) {
	//noinspection SpellCheckingInspection
	return "minecraft:bedrock", map[string]any{"infiniburn_bit": b.InfiniteBurning}
}
