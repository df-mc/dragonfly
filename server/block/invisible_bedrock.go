package block

// InvisibleBedrock is an indestructible, solid block, similar to bedrock and has the appearance of air.
// It shares many of its properties with barriers.
type InvisibleBedrock struct {
	transparent
	solid
}

// EncodeItem ...
func (InvisibleBedrock) EncodeItem() (name string, meta int16) {
	return "minecraft:invisible_bedrock", 0
}

// EncodeBlock ...
func (InvisibleBedrock) EncodeBlock() (string, map[string]any) {
	return "minecraft:invisible_bedrock", nil
}
