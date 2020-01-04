package block

// Bedrock is a block that is indestructible in survival.
type Bedrock struct {
	// InfiniteBurning specifies if the bedrock block is set aflame and will burn forever. This is the case
	// for bedrock found under end crystals on top of the end pillars.
	InfiniteBurning bool
}

func (b Bedrock) Minecraft() (name string, properties map[string]interface{}) {
	return "minecraft:bedrock", map[string]interface{}{"infiniburn_bit": b.InfiniteBurning}
}

func (Bedrock) Name() string {
	return "Bedrock"
}
