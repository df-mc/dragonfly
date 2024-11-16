package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

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
	if s.Type == SmoothSandstone() {
		return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(s)).withBlastResistance(30)
	}
	return newBreakInfo(0.8, pickaxeHarvestable, pickaxeEffective, oneOf(s))
}

// EncodeItem ...
func (s Sandstone) EncodeItem() (name string, meta int16) {
	var prefix string
	if s.Type != NormalSandstone() {
		prefix = s.Type.String() + "_"
	}
	if s.Red {
		return "minecraft:" + prefix + "red_sandstone", 0
	}
	return "minecraft:" + prefix + "sandstone", 0
}

// EncodeBlock ...
func (s Sandstone) EncodeBlock() (string, map[string]any) {
	var prefix string
	if s.Type != NormalSandstone() {
		prefix = s.Type.String() + "_"
	}
	if s.Red {
		return "minecraft:" + prefix + "red_sandstone", nil
	}
	return "minecraft:" + prefix + "sandstone", nil
}

// SmeltInfo ...
func (s Sandstone) SmeltInfo() item.SmeltInfo {
	if s.Type == NormalSandstone() {
		return newSmeltInfo(item.NewStack(Sandstone{Red: s.Red, Type: SmoothSandstone()}, 1), 0.1)
	}
	return item.SmeltInfo{}
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
