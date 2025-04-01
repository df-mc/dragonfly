package block

import "github.com/df-mc/dragonfly/server/item"

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
		return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(s)).withBlastResistance(30)
	}
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, silkTouchOneOf(Cobblestone{}, Stone{})).withBlastResistance(30)
}

// BreakInfo ...
func (g Granite) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(g)).withBlastResistance(30)
}

// BreakInfo ...
func (d Diorite) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBlastResistance(30)
}

// BreakInfo ...
func (a Andesite) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(a)).withBlastResistance(30)
}

// SmeltInfo ...
func (s Stone) SmeltInfo() item.SmeltInfo {
	if s.Smooth {
		return item.SmeltInfo{}
	}
	return newSmeltInfo(item.NewStack(Stone{Smooth: true}, 1), 0.1)
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
	return "minecraft:stone", nil
}

// EncodeItem ...
func (a Andesite) EncodeItem() (name string, meta int16) {
	if a.Polished {
		return "minecraft:polished_andesite", 0
	}
	return "minecraft:andesite", 0
}

// EncodeBlock ...
func (a Andesite) EncodeBlock() (string, map[string]any) {
	if a.Polished {
		return "minecraft:polished_andesite", nil
	}
	return "minecraft:andesite", nil
}

// EncodeItem ...
func (d Diorite) EncodeItem() (name string, meta int16) {
	if d.Polished {
		return "minecraft:polished_diorite", 0
	}
	return "minecraft:diorite", 0
}

// EncodeBlock ...
func (d Diorite) EncodeBlock() (string, map[string]any) {
	if d.Polished {
		return "minecraft:polished_diorite", nil
	}
	return "minecraft:diorite", nil
}

// EncodeItem ...
func (g Granite) EncodeItem() (name string, meta int16) {
	if g.Polished {
		return "minecraft:polished_granite", 0
	}
	return "minecraft:granite", 0
}

// EncodeBlock ...
func (g Granite) EncodeBlock() (string, map[string]any) {
	if g.Polished {
		return "minecraft:polished_granite", nil
	}
	return "minecraft:granite", nil
}
