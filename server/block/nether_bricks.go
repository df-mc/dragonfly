package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// NetherBricks are blocks used to form nether fortresses in the Nether.
// Red Nether bricks, Cracked Nether bricks and Chiseled Nether bricks are decorative variants that do not naturally generate.
type NetherBricks struct {
	solid
	bassDrum

	// NetherBricksType is the type of nether bricks of the block.
	Type NetherBricksType
}

func (n NetherBricks) BreakInfo() BreakInfo {
	return newBreakInfo(2, pickaxeHarvestable, pickaxeEffective, oneOf(n)).withBlastResistance(30)
}

func (n NetherBricks) SmeltInfo() item.SmeltInfo {
	if n.Type == NormalNetherBricks() {
		return newSmeltInfo(item.NewStack(NetherBricks{Type: CrackedNetherBricks()}, 1), 0.1)
	}
	return item.SmeltInfo{}
}

func (n NetherBricks) EncodeItem() (id string, meta int16) {
	return "minecraft:" + n.Type.String(), 0
}

func (n NetherBricks) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + n.Type.String(), nil
}

// allNetherBricks returns a list of all nether bricks variants.
func allNetherBricks() (netherBricks []world.Block) {
	for _, t := range NetherBricksTypes() {
		netherBricks = append(netherBricks, NetherBricks{Type: t})
	}
	return
}
