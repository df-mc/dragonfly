package block

type (
	// Stone is a block found underground in the Overworld or on mountains.
	Stone struct{}
	// Granite is a type of igneous rock.
	Granite polishable
	// Diorite is a type of igneous rock.
	Diorite polishable
	// Andesite is a type of igneous rock.
	Andesite polishable

	// polishable forms the base of blocks that may be polished.
	polishable struct {
		// Polished specifies if the block is polished or not. When set to true, the block will represent its
		// polished variant, for example polished andesite.
		Polished bool
	}
)

func (a Andesite) Minecraft() (name string, properties map[string]interface{}) {
	if a.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "andesite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "andesite"}
}

func (d Diorite) Minecraft() (name string, properties map[string]interface{}) {
	if d.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "diorite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "diorite"}
}

func (g Granite) Minecraft() (name string, properties map[string]interface{}) {
	if g.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "granite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "granite"}
}

func (Stone) Name() string {
	return "Stone"
}

func (Stone) Minecraft() (name string, properties map[string]interface{}) {
	return "minecraft:stone", map[string]interface{}{"stone_type": "stone"}
}

func (g Granite) Name() string {
	if g.Polished {
		return "Polished Granite"
	}
	return "Granite"
}

func (d Diorite) Name() string {
	if d.Polished {
		return "Polished Diorite"
	}
	return "Diorite"
}

func (a Andesite) Name() string {
	if a.Polished {
		return "Polished Andesite"
	}
	return "Andesite"
}
