package block

// Bricks are decorative building blocks.
type Bricks struct {
	solid
	bassDrum
}

// BreakInfo ...
func (b Bricks) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(b))
}

// Resistance ...
func (b Bricks) Resistance() float64 {
	return 6
}

// AlwaysExplodeDrop ..
func (b Bricks) AlwaysExplodeDrop() bool {
	return false
}

// EncodeItem ...
func (Bricks) EncodeItem() (name string, meta int16) {
	return "minecraft:brick_block", 0
}

// EncodeBlock ...
func (Bricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:brick_block", nil
}
