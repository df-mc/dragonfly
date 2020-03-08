package block

type (
	// Stone is a block found underground in the world or on mountains.
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

// EncodeItem ...
func (a Andesite) EncodeItem() (id int32, meta int16) {
	if a.Polished {
		return 1, 6
	}
	return 1, 5
}

// EncodeItem ...
func (d Diorite) EncodeItem() (id int32, meta int16) {
	if d.Polished {
		return 1, 4
	}
	return 1, 3
}

// EncodeItem ...
func (g Granite) EncodeItem() (id int32, meta int16) {
	if g.Polished {
		return 1, 2
	}
	return 1, 1
}

// EncodeItem ...
func (s Stone) EncodeItem() (id int32, meta int16) {
	return 1, 0
}

// EncodeBlock ...
func (a Andesite) EncodeBlock() (name string, properties map[string]interface{}) {
	if a.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "andesite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "andesite"}
}

// EncodeBlock ...
func (d Diorite) EncodeBlock() (name string, properties map[string]interface{}) {
	if d.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "diorite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "diorite"}
}

// EncodeBlock ...
func (g Granite) EncodeBlock() (name string, properties map[string]interface{}) {
	if g.Polished {
		return "minecraft:stone", map[string]interface{}{"stone_type": "granite_smooth"}
	}
	return "minecraft:stone", map[string]interface{}{"stone_type": "granite"}
}

// EncodeBlock ...
func (Stone) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:stone", map[string]interface{}{"stone_type": "stone"}
}
