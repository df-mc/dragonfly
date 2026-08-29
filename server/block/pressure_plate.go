package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// PressurePlate emits redstone power while an entity stands on it.
type PressurePlate struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type is the pressure plate material.
	Type PressurePlateType
	// Power is the current signal strength.
	Power int
}

// UseOnBlock places a pressure plate on a solid surface.
func (p PressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, p)
	if !used || !attachmentSupported(tx, pos, cube.FaceUp) {
		return false
	}
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EntityInside activates the plate for a valid entity.
func (p PressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !p.detects(e) || !entityIntersects(e, pressurePlateActivationBox(pos)) {
		return
	}
	if p.Power > 0 {
		return
	}
	power := 15
	if p.Type.Weighted() {
		power = max(1, p.detectPower(pos, tx))
	}
	p.Power = power
	tx.SetBlock(pos, p, nil)
	tx.ScheduleBlockUpdate(pos, p, p.releaseDelay())
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOn{})
}

// NeighbourUpdateTick breaks an unsupported pressure plate.
func (p PressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !attachmentSupported(tx, pos, cube.FaceUp) {
		breakBlock(p, pos, tx)
	}
}

// ScheduledTick updates the plate's power from the entities on it.
func (p PressurePlate) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	power := p.detectPower(pos, tx)
	if power > 0 {
		if p.Power != power {
			p.Power = power
			tx.SetBlock(pos, p, nil)
		}
		tx.ScheduleBlockUpdate(pos, p, p.releaseDelay())
		return
	}
	if p.Power == 0 {
		return
	}
	p.Power = 0
	tx.SetBlock(pos, p, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.PressurePlateClickOff{})
}

// RedstonePower returns the plate's signal strength.
func (p PressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	return p.Power
}

// RedstoneStrongPower powers the block below the plate.
func (p PressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if face == cube.FaceDown {
		return p.Power
	}
	return 0
}

// BreakInfo returns the plate's break information.
func (p PressurePlate) BreakInfo() BreakInfo {
	effective := pickaxeEffective
	if p.Type.Wood() {
		effective = axeEffective
	}
	return newBreakInfo(0.5, alwaysHarvestable, effective, oneOf(PressurePlate{Type: p.Type}))
}

// SideClosed reports that pressure plates do not close block faces.
func (PressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FuelInfo returns the plate's fuel properties.
func (p PressurePlate) FuelInfo() item.FuelInfo {
	if p.Type.Flammable() {
		return newFuelInfo(time.Second * 15)
	}
	return item.FuelInfo{}
}

// EncodeItem encodes the pressure plate as an item.
func (p PressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + p.Type.String(), 0
}

// EncodeBlock encodes the pressure plate as a block.
func (p PressurePlate) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + p.Type.String(), map[string]any{"redstone_signal": int32(world.ClampRedstonePower(p.Power))}
}

// detects checks whether an entity can activate the plate.
func (p PressurePlate) detects(e world.Entity) bool {
	if player, ok := e.(interface{ GameMode() world.GameMode }); ok && !player.GameMode().HasCollision() {
		return false
	}
	entityType := e.H().Type().EncodeEntity()
	if entityType == "minecraft:xp_orb" {
		return false
	}
	if p.Type.Wood() || p.Type.Weighted() {
		return true
	}
	if living, ok := e.(pressurePlateLivingEntity); ok {
		return !living.Dead()
	}
	return entityType == "minecraft:armor_stand"
}

func (p PressurePlate) entitiesOn(pos cube.Pos, tx *world.Tx, limit int) int {
	box, n := pressurePlateActivationBox(pos), 0
	for e := range tx.EntitiesWithin(box.Grow(1)) {
		if !p.detects(e) || !entityIntersects(e, box) {
			continue
		}
		if n++; n >= limit {
			break
		}
	}
	return n
}

// detectPower calculates the signal produced by entities on the plate.
func (p PressurePlate) detectPower(pos cube.Pos, tx *world.Tx) int {
	switch p.Type {
	case LightWeightedPressurePlate():
		return p.entitiesOn(pos, tx, 15)
	case HeavyWeightedPressurePlate():
		return (p.entitiesOn(pos, tx, 150) + 9) / 10
	}
	if p.entitiesOn(pos, tx, 1) > 0 {
		return 15
	}
	return 0
}

func (p PressurePlate) releaseDelay() time.Duration {
	if p.Type.Weighted() {
		return time.Second / 2
	}
	return time.Second
}

// pressurePlateLivingEntity is implemented by living entities.
type pressurePlateLivingEntity interface {
	Health() float64
	Dead() bool
}

func pressurePlateActivationBox(pos cube.Pos) cube.BBox {
	return cube.Box(0.125, 0, 0.125, 0.875, 0.25, 0.875).Translate(pos.Vec3())
}

func allPressurePlates() (plates []world.Block) {
	for _, t := range PressurePlateTypes() {
		for power := 0; power <= 15; power++ {
			plates = append(plates, PressurePlate{Type: t, Power: power})
		}
	}
	return
}
