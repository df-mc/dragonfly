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
	return newBreakInfo(0.5, harvestTool, effectiveTool, oneOf(p)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateAroundRedstone(pos, tx)
	})
}

// RedstoneSource ...
func (p PressurePlate) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (p PressurePlate) WeakPower(cube.Pos, cube.Face, *world.Tx, bool) int {
	return p.Power
}

// StrongPower ...
func (p PressurePlate) StrongPower(pos cube.Pos, face cube.Face, w *world.Tx, redstone bool) int {
	return p.Power
}

func (p PressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	var power int

	entitySeq := tx.EntitiesWithin(cube.Box(
		float64(pos.X()), float64(pos.Y()), float64(pos.Z()),
		float64(pos.X()+1), float64(pos.Y()+1), float64(pos.Z()+1),
	))

	entityCount := 0
	for range entitySeq {
		entityCount++
	}

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
		tx.PlaySound(pos.Vec3Centre(), sound.PowerOn{})
		tx.SetBlock(pos, p, &world.SetOpts{DisableBlockUpdates: false})
		updateAroundRedstone(pos, tx)
	}

	tx.ScheduleBlockUpdate(pos, p, time.Millisecond*50)
}

// ScheduledTick ...
func (p PressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, r *rand.Rand) {
	if p.Power == 0 {
		return
	}

	entitySeq := tx.EntitiesWithin(cube.Box(
		float64(pos.X()), float64(pos.Y()), float64(pos.Z()),
		float64(pos.X()+1), float64(pos.Y()+1), float64(pos.Z()+1),
	))

	entityCount := 0
	for range entitySeq {
		entityCount++
	}

	if entityCount != 0 {
		return
	}

	p.Power = 0
	tx.SetBlock(pos, p, &world.SetOpts{DisableBlockUpdates: false})
	tx.PlaySound(pos.Vec3Centre(), sound.PowerOff{})
	updateAroundRedstone(pos, tx)
}

// NeighbourUpdateTick ...
func (p PressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if d, ok := tx.Block(pos.Side(cube.FaceDown)).(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		breakBlock(p, pos, tx)
	}
}

// UseOnBlock ...
func (p PressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, p)
	if !used {
		return false
	}

	belowPos := pos.Side(cube.FaceDown)
	if !tx.Block(belowPos).Model().FaceSolid(belowPos, cube.FaceUp, tx) {
		return false
	}

	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// SideClosed ...
func (p PressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
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
