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

// PressurePlate is a non-solid block that emits redstone power while entities
// stand on it. Weighted variants emit an analog power level based on the
// number of entities on the plate.
type PressurePlate struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type is the material the pressure plate is made of.
	Type PressurePlateType
	// Power is the current redstone signal emitted by the plate.
	Power int
}

// UseOnBlock places the pressure plate on a solid surface.
func (p PressurePlate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, p)
	if !used || !attachmentSupported(tx, pos, cube.FaceUp) {
		return false
	}
	place(tx, pos, p, user, ctx)
	return placed(ctx)
}

// EntityInside powers the plate when an entity enters its activation area.
func (p PressurePlate) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if !p.detects(e) || !entityIntersects(e, pressurePlateActivationBox(pos)) {
		return
	}
	if p.Power > 0 {
		// The plate is already active. Its scheduled tick keeps the (weighted)
		// level current, so a stepping entity only needs to defer the release.
		tx.ScheduleBlockUpdate(pos, p, p.releaseDelay())
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

// NeighbourUpdateTick breaks the pressure plate if its supporting block is removed.
func (p PressurePlate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !attachmentSupported(tx, pos, cube.FaceUp) {
		breakBlock(p, pos, tx)
	}
}

// ScheduledTick releases the plate if no entity keeps it pressed.
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

// RedstonePower returns the plate's analog power level.
func (p PressurePlate) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	return p.Power
}

// RedstoneStrongPower strongly powers the block below the pressure plate.
func (p PressurePlate) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if face == cube.FaceDown {
		return p.Power
	}
	return 0
}

// BreakInfo ...
func (p PressurePlate) BreakInfo() BreakInfo {
	effective := pickaxeEffective
	if p.Type.Wood() {
		effective = axeEffective
	}
	return newBreakInfo(0.5, alwaysHarvestable, effective, oneOf(PressurePlate{Type: p.Type}))
}

// SideClosed ...
func (PressurePlate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FuelInfo ...
func (p PressurePlate) FuelInfo() item.FuelInfo {
	if p.Type.Flammable() {
		return newFuelInfo(time.Second * 15)
	}
	return item.FuelInfo{}
}

// EncodeItem ...
func (p PressurePlate) EncodeItem() (name string, meta int16) {
	return "minecraft:" + p.Type.String(), 0
}

// EncodeBlock ...
func (p PressurePlate) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + p.Type.String(), map[string]any{"redstone_signal": int32(world.ClampRedstonePower(p.Power))}
}

// detects reports whether an entity activates the plate. Stone-like plates only
// react to living entities and armour stands; wooden and weighted plates react
// to any entity.
func (p PressurePlate) detects(e world.Entity) bool {
	if p.Type.Wood() || p.Type.Weighted() {
		return true
	}
	if living, ok := e.(pressurePlateLivingEntity); ok {
		return !living.Dead()
	}
	return e.H().Type().EncodeEntity() == "minecraft:armor_stand"
}

// entitiesOn counts the entities intersecting the plate's activation box,
// stopping early once limit is reached.
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

// detectPower returns the power level the entities on the plate produce.
// Weighted plates emit one level per entity, or per ten entities rounded up for
// the heavy variant; every other plate emits full power for any entity at all.
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

// releaseDelay is the delay before the plate re-checks its entities: 0.5
// seconds for weighted plates and 1 second otherwise.
func (p PressurePlate) releaseDelay() time.Duration {
	if p.Type.Weighted() {
		return time.Second / 2
	}
	return time.Second
}

// pressurePlateLivingEntity is implemented by entities that can die. Health is
// part of the interface so that only entities with a full health state match,
// even though Dead alone decides whether the plate reacts.
type pressurePlateLivingEntity interface {
	Health() float64
	Dead() bool
}

// pressurePlateActivationBox is the box entities must intersect to press the
// plate at a position.
func pressurePlateActivationBox(pos cube.Pos) cube.BBox {
	return cube.Box(0.125, 0, 0.125, 0.875, 0.25, 0.875).Translate(pos.Vec3())
}

// allPressurePlates ...
func allPressurePlates() (plates []world.Block) {
	for _, t := range PressurePlateTypes() {
		for power := 0; power <= 15; power++ {
			plates = append(plates, PressurePlate{Type: t, Power: power})
		}
	}
	return
}
