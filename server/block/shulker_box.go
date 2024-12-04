package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type ShulkerBox struct {
	chest

	Type ShulkerBoxType
	Axis cube.Direction
}

func (s ShulkerBox) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, pickaxeEffective, oneOf(s)).withBlastResistance(10)
}

func (s ShulkerBox) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + s.Type.String(), nil // map[string]any{"facing": s.Axis.String()}
}

func (s ShulkerBox) EncodeItem() (id string, meta int16) {
	return "minecraft:" + s.Type.String(), 0
}

func allShulkerBox() (shulkerboxes []world.Block) {
	for _, t := range ShulkerBoxTypes() {
		shulkerboxes = append(shulkerboxes, ShulkerBox{Type: t})
	}

	return
}
