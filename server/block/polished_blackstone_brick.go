package block

type PolishedBlackstoneBrick struct {
	solid
	bassDrum

	// Cracked specifies if the polished blackstone bricks is its cracked variant.
	Cracked bool
}

// BreakInfo ...
func (b PolishedBlackstoneBrick) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(b)).withBlastResistance(6)
}

// EncodeItem ...
func (b PolishedBlackstoneBrick) EncodeItem() (name string, meta int16) {
	name = "polished_blackstone_bricks"
	if b.Cracked {
		name = "cracked_" + name
	}
	return "minecraft:" + name, 0
}

// EncodeBlock ...
func (b PolishedBlackstoneBrick) EncodeBlock() (string, map[string]any) {
	name := "polished_blackstone_bricks"
	if b.Cracked {
		name = "cracked_" + name
	}
	return "minecraft:" + name, nil
}
