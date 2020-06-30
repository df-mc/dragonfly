package item

import (
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

type Flint struct {
	None tool.None
}

// AttackDamage returns the attack damage of the flint.
func (f Flint) AttackDamage() float64 {
	return 1
}

// MaxCount always returns 64.
func (f Flint) MaxCount() int {
	return 64
}

// ToolType ...
func (f Flint) ToolType() tool.Type {
	return f.None.ToolType()
}

// HarvestLevel ...
func (f Flint) HarvestLevel() int {
	return f.None.HarvestLevel()
}

// BaseMiningEfficiency ...
func (f Flint) BaseMiningEfficiency(b world.Block) float64 {
	return f.None.BaseMiningEfficiency(b)
}

// EncodeItem ...
func (f Flint) EncodeItem() (id int32, meta int16) {
	return 318, 0
}
