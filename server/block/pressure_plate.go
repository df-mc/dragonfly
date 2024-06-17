package block

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/particle"
	"github.com/go-gl/mathgl/mgl64"
)

// Lever is a non-solid block that can provide switchable redstone power.
type PressurePlate struct {
	thin
	transparent
	flowingWaterDisplacer

	// Wooden bool
	// Wood   WoodType
	Powered bool
}

// FaceSolid ...
func (p PressurePlate) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}

// Source ...
func (p PressurePlate) Source() bool {
	return true
}

// WeakPower ...
func (p PressurePlate) WeakPower(cube.Pos, cube.Face, *world.World, bool) int {
	if p.Powered {
		return 15
	}
	return 0
}

// StrongPower ...
func (p PressurePlate) StrongPower(_ cube.Pos, face cube.Face, _ *world.World, _ bool) int {
	if p.Powered {
		return 15
	}
	return 0
}

// NeighbourUpdateTick ...
func (p PressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, air := w.Block(pos.Side(cube.FaceDown)).(Air); air {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: Stone{}})
	}
}

// UseOnBlock ...
func (p PressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, p)
	if !used {
		return false
	}

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

func (p PressurePlate) EntityInside(pos cube.Pos, w *world.World, e world.Entity) {
	w.ScheduleBlockUpdate(pos, time.Millisecond*200)
}

func (p PressurePlate) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	bbox := cube.Box(0, 0, 0, 1, 1, 1).Stretch(cube.X, float64(1)/float64(8)).Stretch(cube.Z, float64(1)/float64(8)).ExtendTowards(cube.FaceDown, float64(-3)/float64(4)).Translate(pos.Vec3())
	ent := w.EntitiesWithin(bbox, func(entity world.Entity) bool { return false })

	// NOTE: DO NOT PLAY SOUND HERE
	// I want to implement this on our core-side, so that we can have a "silent-plates" setting.
	p.Powered = false
	if len(ent) >= 1 {
		p.Powered = true
	}

	w.SetBlock(pos, p, nil)
	updateAroundRedstone(pos, w)
}

// BreakInfo ...
func (p PressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.8, pickaxeHarvestable, pickaxeEffective, nil)

}

// EncodeItem ...
func (p PressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:stone_pressure_plate", 0
	// if !p.Wooden {
	// 	return "minecraft:stone_pressure_plate", 0
	// } else {
	// 	w := p.Wood.String()
	// 	if w == "oak" {
	// 		w = "wooden"
	// 	}
	// 	return "minecraft:" + w + "_pressure_plate", 0
	// }
}

// EncodeBlock ...
func (p PressurePlate) EncodeBlock() (string, map[string]any) {
	return "minecraft:stone_pressure_plate", map[string]any{
		"redstone_signal": int32(boolByte(p.Powered)),
	}
	// if !p.Wooden {
	// 	return "minecraft:stone_pressure_plate", map[string]any{
	// 		"redstone_signal": int32(boolByte(p.Powered)),
	// 	}
	// } else {
	// 	w := p.Wood.String()
	// 	if w == "oak" {
	// 		w = "wooden"
	// 	}
	// 	return "minecraft:" + w + "_pressure_plate", map[string]any{
	// 		"redstone_signal": int32(boolByte(p.Powered)),
	// 	}
	// }
}

// allPlanks returns all planks types.
func allPressurePlates() (plates []world.Block) {
	for _, b := range []bool{false, true} {
		plates = append(plates, PressurePlate{Powered: b})
		// for _, w := range WoodTypes() {
		// 	plates = append(plates, PressurePlate{Wooden: true, Wood: w})
		// }
	}
	fmt.Println(plates)
	return
}
