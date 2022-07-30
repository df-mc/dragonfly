package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type RedstoneLamp struct {
	solid
	// Lit is if the redstone lamp is lit and disableEmitting light.
	Lit bool
}

// BreakInfo ...
func (l RedstoneLamp) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, oneOf(l))
}

// LightEmissionLevel ...
func (l RedstoneLamp) LightEmissionLevel() uint8 {
	if l.Lit {
		return 15
	}
	return 0
}

// EncodeItem ...
func (l RedstoneLamp) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_lamp", 0
}

// EncodeBlock ...
func (l RedstoneLamp) EncodeBlock() (string, map[string]any) {
	if l.Lit {
		return "minecraft:lit_redstone_lamp", nil
	}
	return "minecraft:redstone_lamp", nil
}

// UseOnBlock ...
func (l RedstoneLamp) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, l)
	if !used {
		return
	}
	l.Lit = w.ReceivingRedstonePower(pos)
	place(w, pos, l, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (l RedstoneLamp) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if l.Lit != w.ReceivingRedstonePower(pos) {
		l.Lit = !l.Lit
		w.SetBlock(pos, l, nil)
	}
}
