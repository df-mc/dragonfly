package block

// Dirt is a block found abundantly in most biomes under a layer of grass blocks at the top of the
// Overworld.
type Dirt struct {
	// Coarse specifies if the dirt should be off the coarse dirt variant. Grass blocks won't spread on
	// the block if set to true.
	Coarse bool
}

func (d Dirt) EncodeItem() (id int32, meta int16) {
	if d.Coarse {
		return 3, 1
	}
	return 3, 0
}

func (d Dirt) Minecraft() (name string, properties map[string]interface{}) {
	if d.Coarse {
		return "minecraft:dirt", map[string]interface{}{"dirt_type": "coarse"}
	}
	return "minecraft:dirt", map[string]interface{}{"dirt_type": "normal"}
}

func (d Dirt) Name() string {
	if d.Coarse {
		return "Coarse Dirt"
	}
	return "Dirt"
}
