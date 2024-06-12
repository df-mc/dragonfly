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
func (u PressurePlate) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}

// Source ...
func (l PressurePlate) Source() bool {
	return true
}

// WeakPower ...
func (l PressurePlate) WeakPower(cube.Pos, cube.Face, *world.World, bool) int {
	if l.Powered {
		return 15
	}
	return 0
}

// StrongPower ...
func (l PressurePlate) StrongPower(_ cube.Pos, face cube.Face, _ *world.World, _ bool) int {
	if l.Powered {
		return 15
	}
	return 0
}

// NeighbourUpdateTick ...
func (l PressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, air := w.Block(pos.Side(cube.FaceDown)).(Air); air {
		w.SetBlock(pos, nil, nil)
		w.AddParticle(pos.Vec3Centre(), particle.BlockBreak{Block: Stone{}})
	}
}

// UseOnBlock ...
func (l PressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	return true
}

// Activate ...
func (l PressurePlate) Activate(pos cube.Pos, _ cube.Face, w *world.World, _ item.User, _ *item.UseContext) bool {
	fmt.Println("ACTIVATE?")
	return false
}

func (l PressurePlate) EntityInside(pos cube.Pos, w *world.World, e world.Entity) {
	fmt.Println("TEST")
	l.Powered = true
	w.SetBlock(pos, l, nil)
	w.ScheduleBlockUpdate(pos, time.Millisecond*200)
	updateAroundRedstone(pos, w)
}

func (l PressurePlate) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	fmt.Println("TEST")
	bbox := cube.Box(0, 0, 0, 1, 1, 1).Stretch(cube.X, float64(1)/float64(8)).Stretch(cube.Z, float64(1)/float64(8)).ExtendTowards(cube.FaceUp, float64(-3)/float64(4)).Translate(pos.Vec3())
	ent := w.EntitiesWithin(bbox, nil)
	fmt.Println(ent)
	// for _, e := range w.Entities() {
	// 	// What the bullshit
	// 	// TODO: make this by bbox
	// 	p := e.Position()
	// 	if pos.X() == int(math.Floor(p.X())) && pos.Y() == int(math.Floor(p.Y())) && pos.Z() == int(math.Floor(p.Z())) {
	// 		inside = true
	// 	}
	// }
	if len(ent) < 1 {
		l.Powered = false
	}
	w.SetBlock(pos, l, nil)
	//updateAroundRedstone(pos, w)
}

// BreakInfo ...
func (l PressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.8, pickaxeHarvestable, pickaxeEffective, nil)

}

// EncodeItem ...
func (l PressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:stone_pressure_plate", 0
	// if !l.Wooden {
	// 	return "minecraft:stone_pressure_plate", 0
	// } else {
	// 	w := l.Wood.String()
	// 	if w == "oak" {
	// 		w = "wooden"
	// 	}
	// 	return "minecraft:" + w + "_pressure_plate", 0
	// }
}

// EncodeBlock ...
func (l PressurePlate) EncodeBlock() (string, map[string]any) {
	return "minecraft:stone_pressure_plate", map[string]any{
		"redstone_signal": int32(boolByte(l.Powered)),
	}
	// if !l.Wooden {
	// 	return "minecraft:stone_pressure_plate", map[string]any{
	// 		"redstone_signal": int32(boolByte(l.Powered)),
	// 	}
	// } else {
	// 	w := l.Wood.String()
	// 	if w == "oak" {
	// 		w = "wooden"
	// 	}
	// 	return "minecraft:" + w + "_pressure_plate", map[string]any{
	// 		"redstone_signal": int32(boolByte(l.Powered)),
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