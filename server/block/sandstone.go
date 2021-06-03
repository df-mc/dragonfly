package block

import "github.com/df-mc/dragonfly/server/world"

// Sandstone is a solid block commonly found in deserts and beaches underneath sand.
type Sandstone struct {
	solid
	bassDrum

	// Type is the type of sandstone of the block.
	Type SandstoneType

	// Red specifies if the sandstone type is red or not. When set to true, the sandstone type will represent its
	// red variant, for example red sandstone.
	Red bool
}

// BreakInfo ...
func (s Sandstone) BreakInfo() BreakInfo {
	i := newBreakInfo(0.8, pickaxeHarvestable, pickaxeEffective, oneOf(s))
	if s.Type == SmoothSandstone() {
		i.Hardness = 2.0
	}
	return i
}

// EncodeItem ...
func (s Sandstone) EncodeItem() (name string, meta int16) {
	if s.Red {
		return "minecraft:red_sandstone", int16(s.Type.Uint8())
	}
	return "minecraft:sandstone", int16(s.Type.Uint8())
}

// EncodeBlock ...
func (s Sandstone) EncodeBlock() (string, map[string]interface{}) {
	name := "minecraft:sandstone"
	if s.Red {
		name = "minecraft:red_sandstone"
	}
	return name, map[string]interface{}{"sand_stone_type": s.Type.String()}
}

// allSandstones returns a list of all sandstone block variants.
func allSandstones() (c []world.Block) {
	f := func(red bool) {
		for _, t := range SandstoneTypes() {
			c = append(c, Sandstone{Type: t, Red: red})
		}
	}
	f(true)
	f(false)
	return
}
