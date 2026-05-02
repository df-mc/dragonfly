package block

// InfestedDeepslate is a block that hides a silverfish. It looks identical to deepslate.
type InfestedDeepslate struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedDeepslate) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedDeepslate) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_deepslate", 0
}

// EncodeBlock ...
func (i InfestedDeepslate) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_deepslate", nil
}
