package block

// Mud is a decorative block obtained by using a water bottle on a dirt block.
type Mud struct {
	solid
}

// BreakInfo ...
func (m Mud) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(m))
}

// EncodeItem ...
func (Mud) EncodeItem() (name string, meta int16) {
	return "minecraft:mud", 0
}

// EncodeBlock ...
func (Mud) EncodeBlock() (string, map[string]any) {
	return "minecraft:mud", nil
}
