package block

import (
	"math/rand/v2"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"github.com/df-mc/dragonfly/server/world"
)

// Ice is a translucent solid block that melts into water when a bright light source is placed next to it.
type Ice struct {
	solid
}

// Friction ...
func (Ice) Friction() float64 {
	return 0.98
}

// LightDiffusionLevel ...
func (Ice) LightDiffusionLevel() uint8 {
	return 2
}

// RandomTick ...
func (i Ice) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	for _, f := range cube.Faces() {
		if tx.BlockLight(pos.Side(f)) > 11 {
			i.melt(pos, tx)
			return
		}
	}
}

// BreakInfo ...
func (i Ice) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, silkTouchOnlyDrop(i)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, u item.User) {
		if u == nil {
			return
		}
		if gm, ok := u.(interface{ GameMode() world.GameMode }); ok && gm.GameMode().CreativeInventory() {
			return
		}
		held, _ := u.HeldItems()
		if _, ok := held.Enchantment(enchantment.SilkTouch); ok {
			return
		}
		below := pos.Side(cube.FaceDown)
		b := tx.Block(below)
		if _, liquid := b.(world.Liquid); !liquid && !b.Model().FaceSolid(below, cube.FaceUp, tx) {
			return
		}
		i.melt(pos, tx)
	})
}

// melt ...
func (Ice) melt(pos cube.Pos, tx *world.Tx) {
	if tx.World().Dimension().WaterEvaporates() {
		tx.SetBlock(pos, nil, nil)
		return
	}
	tx.SetBlock(pos, Water{Still: true, Depth: 8}, nil)
}

// EncodeItem ...
func (Ice) EncodeItem() (name string, meta int16) {
	return "minecraft:ice", 0
}

// EncodeBlock ...
func (Ice) EncodeBlock() (string, map[string]any) {
	return "minecraft:ice", nil
}
