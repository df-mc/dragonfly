package block

// InvisibleBedrock is an indestructible, solid block, similar to bedrock and has the appearance of air.
// It shares many of its properties with barriers.
type InvisibleBedrock struct {
	transparent
	solid
}

// EncodeItem ...
func (InvisibleBedrock) EncodeItem() (name string, meta int16) {
	return "minecraft:invisiblebedrock", 0
}

// EncodeBlock ...
func (InvisibleBedrock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:invisibleBedrock", nil
}
