package block

type (
	// Sandstone is a solid block commonly found in deserts and beaches underneath sand.
	Sandstone red
	// ChiseledSandstone is a variant of Sandstone. It appears to have hieroglyphs on it.
	ChiseledSandstone red
	// CutSandstone is similar to Sandstone, but it is smooth with some edges carved into it.
	CutSandstone red
	// SmoothSandstone is a more "polished" variant of Sandstone. It is a building and decorative block.
	SmoothSandstone red

	// red forms the base of sandstone that may be red.
	red struct {
		solid
		bassDrum

		// Red specifies if the sandstone type is red or not. When set to true, the sandstone type will represent its
		// red variant, for example red sandstone.
		Red bool
	}
)

var sandStoneBreakInfo = newBreakInfo(0.8, pickaxeHarvestable, pickaxeEffective, nil)

// BreakInfo ...
func (s Sandstone) BreakInfo() BreakInfo {
	i := sandStoneBreakInfo
	i.Drops = oneOf(s)
	return i
}

// BreakInfo ...
func (c ChiseledSandstone) BreakInfo() BreakInfo {
	i := sandStoneBreakInfo
	i.Drops = oneOf(c)
	return i
}

// BreakInfo ...
func (c CutSandstone) BreakInfo() BreakInfo {
	i := sandStoneBreakInfo
	i.Drops = oneOf(c)
	return i
}

// BreakInfo ...
func (s SmoothSandstone) BreakInfo() BreakInfo {
	i := sandStoneBreakInfo
	i.Drops = oneOf(s)
	i.Hardness = 2.0
	return i
}

// EncodeItem ...
func (s Sandstone) EncodeItem() (name string, meta int16) {
	if s.Red {
		return "minecraft:red_sandstone", 0
	}
	return "minecraft:sandstone", 0
}

// EncodeBlock ...
func (s Sandstone) EncodeBlock() (string, map[string]interface{}) {
	if s.Red {
		return "minecraft:red_sandstone", map[string]interface{}{"sand_stone_type": "default"}
	}
	return "minecraft:sandstone", map[string]interface{}{"sand_stone_type": "default"}
}

// EncodeItem ...
func (s ChiseledSandstone) EncodeItem() (name string, meta int16) {
	if s.Red {
		return "minecraft:red_sandstone", 1
	}
	return "minecraft:sandstone", 1
}

// EncodeBlock ...
func (s ChiseledSandstone) EncodeBlock() (string, map[string]interface{}) {
	if s.Red {
		return "minecraft:red_sandstone", map[string]interface{}{"sand_stone_type": "heiroglyphs"}
	}
	return "minecraft:sandstone", map[string]interface{}{"sand_stone_type": "heiroglyphs"}
}

// EncodeItem ...
func (s CutSandstone) EncodeItem() (name string, meta int16) {
	if s.Red {
		return "minecraft:red_sandstone", 2
	}
	return "minecraft:sandstone", 2
}

// EncodeBlock ...
func (s CutSandstone) EncodeBlock() (string, map[string]interface{}) {
	if s.Red {
		return "minecraft:red_sandstone", map[string]interface{}{"sand_stone_type": "cut"}
	}
	return "minecraft:sandstone", map[string]interface{}{"sand_stone_type": "cut"}
}

// EncodeItem ...
func (s SmoothSandstone) EncodeItem() (name string, meta int16) {
	if s.Red {
		return "minecraft:red_sandstone", 3
	}
	return "minecraft:sandstone", 3
}

// EncodeBlock ...
func (s SmoothSandstone) EncodeBlock() (string, map[string]interface{}) {
	if s.Red {
		return "minecraft:red_sandstone", map[string]interface{}{"sand_stone_type": "smooth"}
	}
	return "minecraft:sandstone", map[string]interface{}{"sand_stone_type": "smooth"}
}
