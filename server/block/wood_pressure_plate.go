package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// WoodPressurePlate is a non-solid block that can detect entities standing on it and provide redstone power.
type WoodPressurePlate struct {
	transparent
	empty
	sourceWaterDisplacer

	// Wood is the type of wood of the pressure plate.
	Wood WoodType
	// Powered reports whether the pressure plate is currently being pressed.
	Powered bool
}

// RedstoneSource ...
func (WoodPressurePlate) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (p WoodPressurePlate) WeakPower(_ cube.Pos, _ cube.Face, _ *world.Tx, _ bool) int {
	if p.Powered {
		return 15
	}
	return 0
}

// StrongPower ...
func (p WoodPressurePlate) StrongPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if p.Powered && face == cube.FaceDown {
		return 15
	}
	return 0
}

// SideClosed ...
func (WoodPressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// NeighbourUpdateTick ...
func (p WoodPressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Air); ok {
		breakBlock(p, pos, tx)
	}
}

// UseOnBlock places the pressure plate on the top face of a block.
func (p WoodPressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, p)
	if !used {
		return false
	}
	if face != cube.FaceUp {
		return false
	}
	supportPos := pos.Side(cube.FaceDown)
	if _, ok := tx.Block(supportPos).(Air); ok {
		return false
	}
	p.Powered = false
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EntityLand is called when an entity lands on the pressure plate, powering it.
func (p WoodPressurePlate) EntityLand(pos cube.Pos, tx *world.Tx, _ world.Entity, _ *float64) {
	if p.Powered {
		return
	}
	p.Powered = true
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PowerOn{})
	tx.ScheduleBlockUpdate(pos, p, time.Millisecond*500)
	updateDirectionalRedstone(pos, tx, cube.FaceDown)
}

// ScheduledTick releases the pressure plate if no entities are on it.
func (p WoodPressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !p.Powered {
		return
	}
	p.Powered = false
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PowerOff{})
	updateDirectionalRedstone(pos, tx, cube.FaceDown)
}

// BreakInfo ...
func (p WoodPressurePlate) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, axeEffective, oneOf(p)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		if p.Powered {
			updateDirectionalRedstone(pos, tx, cube.FaceDown)
		}
	})
}

// FlammabilityInfo ...
func (p WoodPressurePlate) FlammabilityInfo() FlammabilityInfo {
	if !p.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// FuelInfo ...
func (p WoodPressurePlate) FuelInfo() item.FuelInfo {
	if !p.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 10)
}

// EncodeItem ...
func (p WoodPressurePlate) EncodeItem() (name string, meta int16) {
	if p.Wood == OakWood() {
		return "minecraft:wooden_pressure_plate", 0
	}
	return "minecraft:" + p.Wood.String() + "_pressure_plate", 0
}

// EncodeBlock ...
func (p WoodPressurePlate) EncodeBlock() (name string, properties map[string]any) {
	redstoneSignal := int32(0)
	if p.Powered {
		redstoneSignal = 15
	}
	if p.Wood == OakWood() {
		return "minecraft:wooden_pressure_plate", map[string]any{"redstone_signal": redstoneSignal}
	}
	return "minecraft:" + p.Wood.String() + "_pressure_plate", map[string]any{"redstone_signal": redstoneSignal}
}

// allWoodPressurePlates returns a list of all wood pressure plate types.
func allWoodPressurePlates() (plates []world.Block) {
	for _, w := range WoodTypes() {
		plates = append(plates, WoodPressurePlate{Wood: w})
		plates = append(plates, WoodPressurePlate{Wood: w, Powered: true})
	}
	return
}
