package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// PressurePlate is a non-solid block that produces a redstone signal when stood on by an entity
type PressurePlate struct {
	sourceWaterDisplacer
	transparent
	empty

	// Type is the type of the pressure plate.
	Type PressurePlateType
	// Power specifies the redstone power level currently being produced by the pressure plate.
	Power int
}

// BreakInfo ...
func (p PressurePlate) BreakInfo() BreakInfo {
	harvestTool := alwaysHarvestable
	effectiveTool := axeEffective
	if p.Type == StonePressurePlate() || p.Type == PolishedBlackstonePressurePlate() || p.Type == HeavyWeightedPressurePlate() || p.Type == LightWeightedPressurePlate() {
		harvestTool = pickaxeHarvestable
		effectiveTool = pickaxeEffective
	}
	return newBreakInfo(0.5, harvestTool, effectiveTool, oneOf(p)).withBreakHandler(func(pos cube.Pos, w *world.World, _ item.User) {
		updateAroundRedstone(pos, w)
	})
}

// Source ...
func (p PressurePlate) Source() bool {
	return true
}

// WeakPower ...
func (p PressurePlate) WeakPower(cube.Pos, cube.Face, *world.World, bool) int {
	return p.Power
}

// StrongPower ...
func (p PressurePlate) StrongPower(pos cube.Pos, face cube.Face, w *world.World, redstone bool) int {
	return p.Power
}

func (p PressurePlate) EntityInside(pos cube.Pos, w *world.World, e world.Entity) {
	var power int
	entityCount := len(w.EntitiesWithin(cube.Box(
		float64(pos.X()), float64(pos.Y()), float64(pos.Z()),
		float64(pos.X()+1), float64(pos.Y()+1), float64(pos.Z()+1),
	), nil))

	switch p.Type {
	case StonePressurePlate(), PolishedBlackstonePressurePlate():
		//TODO: add a check if its a living entity currently not possible due to import cycle
		power = 15
	case HeavyWeightedPressurePlate():
		power = min(entityCount, 15)
	case LightWeightedPressurePlate():
		power = min((entityCount+9)/10, 15)
	default:
		power = 15
	}

	if power > 0 && power != p.Power {
		p.Power = power
		w.PlaySound(pos.Vec3Centre(), sound.PowerOn{})
		w.SetBlock(pos, p, &world.SetOpts{DisableBlockUpdates: false})
		updateAroundRedstone(pos, w)
	}

	w.ScheduleBlockUpdate(pos, time.Millisecond*50)
}

// ScheduledTick ...
func (p PressurePlate) ScheduledTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if p.Power == 0 {
		return
	}

	entityCount := len(w.EntitiesWithin(cube.Box(float64(pos.X()), float64(pos.Y()), float64(pos.Z()), float64(pos.X()+1), float64(pos.Y()+1), float64(pos.Z()+1)), nil))
	if entityCount != 0 {
		return
	}

	p.Power = 0
	w.SetBlock(pos, p, &world.SetOpts{DisableBlockUpdates: false})
	w.PlaySound(pos.Vec3Centre(), sound.PowerOff{})
	updateAroundRedstone(pos, w)
}

// NeighbourUpdateTick ...
func (p PressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d, ok := w.Block(pos.Side(cube.FaceDown)).(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		w.SetBlock(pos, nil, nil)
		dropItem(w, item.NewStack(p, 1), pos.Vec3Centre())
	}
}

// UseOnBlock ...
func (p PressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, p)
	if !used {
		return false
	}

	belowPos := pos.Side(cube.FaceDown)
	if !w.Block(belowPos).Model().FaceSolid(belowPos, cube.FaceUp, w) {
		return false
	}

	place(w, pos, p, user, ctx)
	return placed(ctx)
}

// SideClosed ...
func (p PressurePlate) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (p PressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + p.Type.String() + "_pressure_plate", 0
}

// EncodeBlock ...
func (p PressurePlate) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + p.Type.String() + "_pressure_plate", map[string]any{"redstone_signal": int32(p.Power)}
}

// allPressurePlates ...
func allPressurePlates() (pressureplates []world.Block) {
	for _, w := range PressurePlateTypes() {
		for i := 0; i <= 15; i++ {
			pressureplates = append(pressureplates, PressurePlate{Type: w, Power: i})
		}
	}
	return
}
