package block

type (
	// Stone is a block found underground in the world or on mountains.
	Stone struct {
		solid
		bassDrum

		// Smooth specifies if the stone is its smooth variant.
		Smooth bool
	}

	// Granite is a type of igneous rock.
	Granite polishable
	// Diorite is a type of igneous rock.
	Diorite polishable
	// Andesite is a type of igneous rock.
	Andesite polishable

	// polishable forms the base of blocks that may be polished.
	polishable struct {
		solid
		bassDrum
		// Polished specifies if the block is polished or not. When set to true, the block will represent its
		// polished variant, for example polished andesite.
		Polished bool
	}
)

// BreakInfo ...
func (s Stone) BreakInfo() BreakInfo {
	if s.Smooth {
		return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(s))
	}
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(Cobblestone{}, Stone{}))
}

// BreakInfo ...
func (g Granite) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(g))
}

// BreakInfo ...
func (d Diorite) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(d))
}

// BreakInfo ...
func (a Andesite) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(a))
}

// EncodeItem ...
func (s Stone) EncodeItem() (name string, meta int16) {
	if s.Smooth {
		return "minecraft:smooth_stone", 0
	}
	return "minecraft:stone", 0
}

// EncodeBlock ...
func (s Stone) EncodeBlock() (string, map[string]any) {
	if s.Smooth {
		return "minecraft:smooth_stone", nil
	}
	return "minecraft:stone", map[string]any{"stone_type": "stone"}
}

// EncodeItem ...
func (a Andesite) EncodeItem() (name string, meta int16) {
	if a.Polished {
		return "minecraft:stone", 6
	}
	return "minecraft:stone", 5
}

// EncodeBlock ...
func (a Andesite) EncodeBlock() (string, map[string]any) {
	if a.Polished {
		return "minecraft:stone", map[string]any{"stone_type": "andesite_smooth"}
	}
	return "minecraft:stone", map[string]any{"stone_type": "andesite"}
}

// EncodeItem ...
func (d Diorite) EncodeItem() (name string, meta int16) {
	if d.Polished {
		return "minecraft:stone", 4
	}
	return "minecraft:stone", 3
}

// EncodeBlock ...
func (d Diorite) EncodeBlock() (string, map[string]any) {
	if d.Polished {
		return "minecraft:stone", map[string]any{"stone_type": "diorite_smooth"}
	}
	return "minecraft:stone", map[string]any{"stone_type": "diorite"}
}

// EncodeItem ...
func (g Granite) EncodeItem() (name string, meta int16) {
	if g.Polished {
		return "minecraft:stone", 2
	}
	return "minecraft:stone", 1
}

// EncodeBlock ...
func (g Granite) EncodeBlock() (string, map[string]any) {
	if g.Polished {
		return "minecraft:stone", map[string]any{"stone_type": "granite_smooth"}
	}
	return "minecraft:stone", map[string]any{"stone_type": "granite"}
}
