package block

// RedstoneBlock is a precious mineral block made from 9 redstone.
type RedstoneBlock struct {
	solid
	bassDrum
}

// BreakInfo ...
func (r RedstoneBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(r)).withBlastResistance(30)
}

// EncodeItem ...
func (RedstoneBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_block", 0
}

// EncodeBlock ...
func (RedstoneBlock) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:redstone_block", nil
}
