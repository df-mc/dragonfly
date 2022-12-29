package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
)

// Ice is a transparent block that forms when water freezes and melts when it is near a bright light source.
type Ice struct {
	solid
}

// LightDiffusionLevel ...
func (Ice) LightDiffusionLevel() uint8 {
	return 3
}

// BreakInfo ...
func (i Ice) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, silkTouchOnlyDrop(i)).
		withBreakHandler(func(pos cube.Pos, w *world.World, u item.User) {
			if w.Dimension().WaterEvaporates() {
				return
			}
			if p, ok := u.(interface {
				GameMode() world.GameMode
			}); ok && p.GameMode().CreativeInventory() {
				return
			}
			if mainHand, _ := u.HeldItems(); hasSilkTouch(mainHand.Enchantments()) {
				return
			}
			if _, ok := w.Block(pos.Side(cube.FaceDown)).Model().(model.Solid); !ok {
				return
			}
			w.SetBlock(pos, Water{Depth: 8}, nil)
		})
}

// RandomTick ...
func (i Ice) RandomTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if w.Light(pos) < 12 {
		return
	}
	if w.Dimension().WaterEvaporates() {
		w.SetBlock(pos, nil, nil)
		return
	}
	w.SetBlock(pos, Water{Depth: 8}, nil)
}

// Friction ...
func (i Ice) Friction() float64 {
	return 0.98
}

// EncodeItem ...
func (Ice) EncodeItem() (name string, meta int16) {
	return "minecraft:ice", 0
}

// EncodeBlock ...
func (Ice) EncodeBlock() (string, map[string]any) {
	return "minecraft:ice", nil
}
